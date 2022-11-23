// Package client manages connections and interactions with the EVM chain
package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

// client wraps the standard Ethereum client
type Wallet struct {
	EthClient *ethclient.Client

	privateKey *ecdsa.PrivateKey
	address    common.Address

	pendingNonce uint64
	nonceMutex   sync.Mutex
}

// NewWallet produces a new wallet to manage transactions on the specified chain
func NewWallet(privateKey *ecdsa.PrivateKey, wsURL string) (*Wallet, error) {
	address, err := PrivateKeyToAddress(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyString := PrivateKeyString(privateKey)
	if wallet, exists := globalWalletManager.isExistingWallet(privateKeyString); exists {
		log.Debug().Str("Private Key", privateKeyString).Msg("Wallet already exists")
		return wallet, nil
	}

	log.Debug().
		Str("Private Key", privateKeyString).
		Str("Address", address.Hex()).
		Str("URL", wsURL).
		Msg("Creating New Wallet")
	c, err := ethclient.Dial(wsURL)
	if err != nil {
		return nil, err
	}

	nonce, err := c.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, err
	}
	wallet := &Wallet{
		EthClient:    c,
		pendingNonce: nonce,
		privateKey:   privateKey,
		address:      address,
	}
	globalWalletManager.addNewWallet(privateKeyString, wallet)
	return wallet, nil
}

// SubscribeNewBlocks wraps SubscribeNewHead
func (w *Wallet) SubscribeNewBlocks(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return w.EthClient.SubscribeNewHead(ctx, ch)
}

// SendTransaction sends an eth transaction, and confirms it gets on chain.
// timeBeforeResending determines how long to wait for the transaction to be confirmed before attempting a re-send
// reSendTipMultiplier multiplies the gas tip when replacing the old transaction
func (w *Wallet) SendTransaction(
	timeBeforeResending time.Duration,
	reSendTipMultiplier *big.Float,
	toAddress common.Address,
	additionalTip *big.Int,
	amount *big.Float,
) {
	if additionalTip.Cmp(big.NewInt(0)) <= 0 {
		additionalTip.SetInt64(1)
	}
	nonce := w.newNonce()
	attempt := 0
	reSendTimer := time.NewTimer(0)
	confirmedChan, errChan := make(chan struct{}), make(chan error)

	var (
		tx  *types.Transaction
		err error
	)
	for {
		select {
		case <-reSendTimer.C: // Transaction took too long to confirm, bump gas and go again!
			reSendTimer = time.NewTimer(timeBeforeResending)
			attempt++
			if attempt > 1 { // If attempting more than once, keep bumping our tip
				additionalTipFloat := big.NewFloat(0).SetUint64(additionalTip.Uint64())
				additionalTipUint, _ := additionalTipFloat.Mul(additionalTipFloat, reSendTipMultiplier).Uint64()
				additionalTip.SetUint64(additionalTipUint)
			}

			tx, err = w.sendTx(toAddress, additionalTip, amount, nonce, attempt)
			if err != nil {
				log.Fatal().Err(err).
					Str("To", toAddress.Hex()).
					Str("From", w.address.Hex()).
					Str("Amount", amount.String()).
					Str("Hash", tx.Hash().Hex()).
					Uint64("Gas Tip Cap", tx.GasTipCap().Uint64()).
					Uint64("Gas Fee Cap", tx.GasFeeCap().Uint64()).
					Uint64("Nonce", nonce).
					Int("Attempt", attempt).
					Msg("Error Sending Transaction")
			}
			go w.ConfirmTx(confirmedChan, errChan, tx)
		case err := <-errChan:
			log.Fatal().Err(err).
				Str("To", toAddress.Hex()).
				Str("From", w.address.Hex()).
				Str("Amount", amount.String()).
				Str("Hash", tx.Hash().Hex()).
				Uint64("Gas Tip Cap", tx.GasTipCap().Uint64()).
				Uint64("Gas Fee Cap", tx.GasFeeCap().Uint64()).
				Uint64("Nonce", nonce).
				Int("Attempt", attempt).
				Msg("Error Confirming Transaction")
		case <-confirmedChan:
			log.Debug().
				Str("To", toAddress.Hex()).
				Str("From", w.address.Hex()).
				Str("Amount", amount.String()).
				Str("Hash", tx.Hash().Hex()).
				Uint64("Nonce", nonce).
				Int("Attempt", attempt).
				Msg("Confirmed Transaction")
			close(errChan)
			close(confirmedChan)
			return
		}
	}
}

func (w *Wallet) sendTx(
	toAddress common.Address,
	additionalTip *big.Int,
	amount *big.Float,
	nonce uint64,
	attempt int,
) (transaction *types.Transaction, err error) {
	fromAddr, err := PrivateKeyToAddress(w.privateKey)
	if err != nil {
		return nil, err
	}
	suggestedGasTipCap, err := w.EthClient.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}

	// Bump Tip Cap
	suggestedGasTipCap.Add(suggestedGasTipCap, additionalTip)

	latestBlock, err := w.EthClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	baseFeeMult := big.NewInt(1).Mul(latestBlock.BaseFee(), big.NewInt(2))
	gasFeeCap := baseFeeMult.Add(baseFeeMult, suggestedGasTipCap)

	tx, err := types.SignNewTx(w.privateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     nonce,
		To:        &toAddress,
		Value:     EtherToWei(amount),
		GasTipCap: suggestedGasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       21000,
	})
	if err != nil {
		return nil, err
	}
	log.Debug().
		Str("From", fromAddr.Hex()).
		Str("To", toAddress.Hex()).
		Str("Amount", amount.String()).
		Str("Hash", tx.Hash().Hex()).
		Uint64("Gas Tip Cap", suggestedGasTipCap.Uint64()).
		Uint64("Gas Fee Cap", gasFeeCap.Uint64()).
		Uint64("Nonce", nonce).
		Int("Attempt", attempt).
		Msg("Sending Transaction")
	return tx, w.EthClient.SendTransaction(context.Background(), tx)
}

// ConfirmTx attempts to confirm a tx on chain, sending a struct on the confirmed chan when it does
func (w *Wallet) ConfirmTx(
	confirmedChan chan struct{},
	errChan chan error,
	transaction *types.Transaction,
) {
	_, isPending, err := w.EthClient.TransactionByHash(context.Background(), transaction.Hash())
	if err != nil {
		errChan <- err
		return
	}
	if !isPending {
		confirmedChan <- struct{}{}
		return
	}
	newBlocks := make(chan *types.Header)
	sub, err := w.EthClient.SubscribeNewHead(context.Background(), newBlocks)
	if err != nil {
		errChan <- err
		return
	}

	for {
		select {
		case err = <-sub.Err():
			errChan <- err
			return
		case <-newBlocks:
			_, isPending, err = w.EthClient.TransactionByHash(context.Background(), transaction.Hash())
			if err != nil {
				errChan <- err
				return
			} else if !isPending { // Confirmed on chain
				confirmedChan <- struct{}{}
				return
			}
		}
	}
}

// ConfirmTxWait confirms the transaction, waiting until the tx confirms until it returns
func (w *Wallet) ConfirmTxWait(ctx context.Context, transaction *types.Transaction) error {
	_, isPending, err := w.EthClient.TransactionByHash(ctx, transaction.Hash())
	if err != nil {
		return err
	}
	if !isPending {
		return nil
	}
	newBlocks := make(chan *types.Header)
	sub, err := w.EthClient.SubscribeNewHead(ctx, newBlocks)
	if err != nil {
		return err
	}

	for {
		select {
		case err = <-sub.Err():
			return err
		case <-ctx.Done():
			return fmt.Errorf("timed out while confirming transaction %s", transaction.Hash().Hex())
		case <-newBlocks:
			_, isPending, err = w.EthClient.TransactionByHash(ctx, transaction.Hash())
			if err != nil {
				return err
			} else if !isPending { // Confirmed on chain
				return nil
			}
		}
	}
}

// TransactionOpts for deploying contracts. Warning: increases nonce
func (w *Wallet) TransactionOpts() (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(w.privateKey, config.Current.BigChainID)
	if err != nil {
		return nil, err
	}
	opts.From = w.address
	opts.Context = context.Background()
	opts.Nonce = new(big.Int).SetUint64(w.newNonce())
	return opts, nil
}

// newNonce gives a new nonce to use for this wallet
func (w *Wallet) newNonce() uint64 {
	w.nonceMutex.Lock()
	defer w.nonceMutex.Unlock()
	nonce := w.pendingNonce
	w.pendingNonce++
	return nonce
}

// Balance retrieves the current balance of the wallet
func (w *Wallet) Balance() (*big.Int, error) {
	bal, err := w.EthClient.BalanceAt(context.Background(), w.address, nil)
	if err != nil {
		return nil, err
	}
	balFloat, _ := WeiToEther(bal).Float64()
	log.Debug().Str("Address", w.address.Hex()).Float64("ETH", balFloat).Msg("Balance")
	return bal, err
}

// BalanceAt wraps balance at
func (w *Wallet) BalanceAt(address common.Address) (*big.Int, error) {
	bal, err := w.EthClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	balFloat, _ := WeiToEther(bal).Float64()
	log.Debug().Str("Address", address.Hex()).Float64("ETH", balFloat).Msg("Balance")
	return bal, err
}

// BlockByNumber wraps ether block by number. Input nil for latest block
func (w *Wallet) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	return w.EthClient.BlockByNumber(context.Background(), blockNumber)
}
