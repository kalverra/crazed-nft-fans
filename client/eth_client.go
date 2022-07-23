// Package client manages connections and interactions with the EVM chain
package client

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/rs/zerolog/log"
)

// EthClient wraps the standard Ethereum client
type EthClient struct {
	innerClient *ethclient.Client
	FundedKey   ecdsa.PrivateKey
}

// NewClient produces a new client to connect to the blockchain
func NewClient(wsURL string) (*EthClient, error) {
	log.Debug().Str("URL", wsURL).Msg("Connecting Client")
	ethClient, err := ethclient.Dial(wsURL)
	return &EthClient{
		innerClient: ethClient,
	}, err
}

// SuggestNonce suggests a nonce to use for sending a transaction
func (c *EthClient) SuggestNonce(account common.Address) (uint64, error) {
	return c.innerClient.PendingNonceAt(context.Background(), account)
}

// SendTransaction sends an eth transaction
func (c *EthClient) SendTransaction(
	privateKey *ecdsa.PrivateKey,
	toAddress common.Address,
	nonce uint64,
	additionalTip *big.Int,
	amount *big.Float,
) (txHash common.Hash, err error) {
	fromAddr, err := PrivateKeyToAddress(privateKey)
	if err != nil {
		return common.Hash{}, err
	}
	suggestedGasTipCap, err := c.innerClient.SuggestGasTipCap(context.Background())
	if err != nil {
		return common.Hash{}, err
	}

	// Bump Tip Cap
	suggestedGasTipCap.Add(suggestedGasTipCap, additionalTip)

	latestBlock, err := c.innerClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return common.Hash{}, err
	}
	baseFeeMult := big.NewInt(1).Mul(latestBlock.BaseFee(), big.NewInt(2))
	gasFeeCap := baseFeeMult.Add(baseFeeMult, suggestedGasTipCap)

	tx, err := types.SignNewTx(privateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     nonce,
		To:        &toAddress,
		Value:     EtherToWei(amount),
		GasTipCap: suggestedGasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       21000,
	})
	if err != nil {
		return common.Hash{}, err
	}

	log.Info().
		Str("From", fromAddr.Hex()).
		Str("To", toAddress.Hex()).
		Str("Amount", amount.String()).
		Uint64("Gas Tip Cap", suggestedGasTipCap.Uint64()).
		Uint64("Gas Fee Cap", gasFeeCap.Uint64()).
		Uint64("Nonce", nonce).
		Msg("Sending Transaction")
	return tx.Hash(), c.innerClient.SendTransaction(context.Background(), tx)
}

// ConfirmTransaction attempts to confirm a pending transaction until the context runs out
func (c *EthClient) ConfirmTransaction(ctxt context.Context, txHash common.Hash) (confirmed bool, err error) {
	_, isPending, err := c.innerClient.TransactionByHash(context.Background(), txHash)
	if !isPending {
		return isPending, err
	}
	newBlocks := make(chan *types.Header)
	sub, err := c.innerClient.SubscribeNewHead(context.Background(), newBlocks)
	if err != nil {
		return isPending, err
	}

	for {
		select {
		case err := <-sub.Err():
			return false, err
		case <-ctxt.Done():
			return false, nil
		case <-newBlocks:
			_, isPending, err = c.innerClient.TransactionByHash(context.Background(), txHash)
			if err != nil {
				return false, err
			}
			if !isPending {
				return true, err
			}
		}
	}
}

// BalanceAt retrieves the current balance of the supplied address
func (c *EthClient) BalanceAt(address common.Address) (*big.Int, error) {
	bal, err := c.innerClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	balFloat, _ := WeiToEther(bal).Float64()
	log.Debug().Str("Address", address.Hex()).Float64("ETH", balFloat).Msg("Balance")
	return bal, err
}
