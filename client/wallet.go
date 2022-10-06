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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

// client wraps the standard Ethereum client
type Wallet struct {
	client *ethclient.Client

	privateKey *ecdsa.PrivateKey
	address    common.Address

	pendingNonce uint64
	nonceMutex   sync.Mutex
}

// NewClient produces a new client connected to the chain provided in the config
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
		Str("Private Key", fmt.Sprintf("%x", privateKeyString)).
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
		client:       c,
		pendingNonce: nonce,
		privateKey:   privateKey,
		address:      address,
	}
	globalWalletManager.addNewWallet(privateKeyString, wallet)
	return wallet, nil
}

// SubscribeNewBlocks wraps SubscribeNewHead
func (w *Wallet) SubscribeNewBlocks(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return w.client.SubscribeNewHead(ctx, ch)
}

// SendTransaction sends an eth transaction
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
	w.nonceMutex.Lock()
	nonce := w.pendingNonce
	w.pendingNonce++
	w.nonceMutex.Unlock()
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
					Str("Amount", amount.String()).
					Str("Hash", tx.Hash().Hex()).
					Uint64("Gas Tip Cap", tx.GasTipCap().Uint64()).
					Uint64("Gas Fee Cap", tx.GasFeeCap().Uint64()).
					Uint64("Nonce", nonce).
					Int("Attempt", attempt).
					Msg("Error Sending Transaction")
			}
			go w.confirmTx(confirmedChan, errChan, tx)
		case err := <-errChan:
			log.Fatal().Err(err).
				Str("To", toAddress.Hex()).
				Str("Amount", amount.String()).
				Str("Hash", tx.Hash().Hex()).
				Uint64("Gas Tip Cap", tx.GasTipCap().Uint64()).
				Uint64("Gas Fee Cap", tx.GasFeeCap().Uint64()).
				Uint64("Nonce", nonce).
				Int("Attempt", attempt).
				Msg("Error Confirming Transaction")
		case <-confirmedChan:
			log.Info().
				Str("To", toAddress.Hex()).
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
	suggestedGasTipCap, err := w.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}

	// Bump Tip Cap
	suggestedGasTipCap.Add(suggestedGasTipCap, additionalTip)

	latestBlock, err := w.client.BlockByNumber(context.Background(), nil)
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
	log.Info().
		Str("From", fromAddr.Hex()).
		Str("To", toAddress.Hex()).
		Str("Amount", amount.String()).
		Str("Hash", tx.Hash().Hex()).
		Uint64("Gas Tip Cap", suggestedGasTipCap.Uint64()).
		Uint64("Gas Fee Cap", gasFeeCap.Uint64()).
		Uint64("Nonce", nonce).
		Int("Attempt", attempt).
		Msg("Sending Transaction")
	return tx, w.client.SendTransaction(context.Background(), tx)
}

// ConfirmTransaction attempts to confirm a pending transaction until the context runs out
func (w *Wallet) confirmTx(
	confirmedChan chan struct{},
	errChan chan error,
	transaction *types.Transaction,
) {
	_, isPending, err := w.client.TransactionByHash(context.Background(), transaction.Hash())
	if err != nil {
		errChan <- err
		return
	}
	if !isPending {
		confirmedChan <- struct{}{}
		return
	}
	newBlocks := make(chan *types.Header)
	sub, err := w.client.SubscribeNewHead(context.Background(), newBlocks)
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
			_, isPending, err = w.client.TransactionByHash(context.Background(), transaction.Hash())
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

// Balance retrieves the current balance of the wallet
func (w *Wallet) Balance() (*big.Int, error) {
	bal, err := w.client.BalanceAt(context.Background(), w.address, nil)
	if err != nil {
		return nil, err
	}
	balFloat, _ := WeiToEther(bal).Float64()
	log.Debug().Str("Address", w.address.Hex()).Float64("ETH", balFloat).Msg("Balance")
	return bal, err
}

// BalanceAt wraps balance at
func (w *Wallet) BalanceAt(address common.Address) (*big.Int, error) {
	bal, err := w.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	balFloat, _ := WeiToEther(bal).Float64()
	log.Debug().Str("Address", address.Hex()).Float64("ETH", balFloat).Msg("Balance")
	return bal, err
}
