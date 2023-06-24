package fans

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/rs/zerolog/log"
)

// TrackedTransaction is a transaction that is being tracked by a wallet and can be resent if it times out
type TrackedTransaction struct {
	Transaction *types.Transaction
	timeSent    time.Time
}

// Wallet manages transactions and re-sends them if they time out
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address

	fan       *Fan
	president *President

	client              *ethclient.Client
	pendingNonce        uint64
	pendingMu           sync.Mutex
	pendingTransactions []*TrackedTransaction
	crazedLevel         *config.CrazedLevel
}

// NewFanWallet creates a new wallet designed to be used by a fan
func NewFanWallet(fan *Fan) (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	pendingNonce, err := fan.client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		return nil, err
	}
	return &Wallet{
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),

		fan: fan,

		client:              fan.client,
		pendingNonce:        pendingNonce,
		pendingTransactions: []*TrackedTransaction{},
		crazedLevel:         fan.CrazedLevel,
	}, nil
}

// NewPresidentWallet creates a new wallet designed to be used by a president
func NewPresidentWallet(president *President) (*Wallet, error) {
	pendingNonce, err := president.Client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		return nil, err
	}
	return &Wallet{
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),

		president: &Pr,

		client:              client,
		pendingNonce:        pendingNonce,
		pendingTransactions: []*TrackedTransaction{},
		crazedLevel:         crazedLevel,
	}, nil
}

// UpdatePendingTxs updates the wallet's pending transactions, removing mined transactions and resending timed out transactions
func (w *Wallet) UpdatePendingTxs(latestBlock *types.Block) {
	for index, tx := range w.pendingTransactions {
		if latestBlock.Transaction(tx.Transaction.Hash()) != nil { // transaction was mined
			w.pendingTransactions = append(w.pendingTransactions[:index], w.pendingTransactions[index+1:]...)
			continue
		}
		if time.Since(tx.timeSent) > w.crazedLevel.TransactionTimeout { // transaction timed out, resend
			w.pendingTransactions = append(w.pendingTransactions[:index], w.pendingTransactions[index+1:]...)
			oldTip := big.NewFloat(0).SetInt64(tx.Transaction.GasTipCap().Int64())
			newGasTip := oldTip.Mul(oldTip, big.NewFloat(w.crazedLevel.GasPriceMultiplier))
			newGasUint, _ := newGasTip.Uint64()
			w.reSendTx(tx.Transaction.Nonce(), big.NewInt(0).SetUint64(newGasUint))
		}
	}
}

// PendingTransactions returns the wallet's currently pending transactions
func (w *Wallet) PendingTransactions() []*TrackedTransaction {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	return w.pendingTransactions
}

// SendTransaction sends a transaction to the chain and tracks and re-sends it if it times out
func (w *Wallet) SendTransaction(latestBlock *types.Block, to common.Address, value *big.Int) error {
	gasUnits, baseFee, gasTipCapFloat, err := w.estimateGas(latestBlock, to)
	if err != nil {
		return err
	}
	gasTipCap := new(big.Int)
	gasTipCapFloat.Int(gasTipCap)
	gasFeeCap := big.NewInt(0).Add(baseFee, gasTipCap)

	tx, err := types.SignNewTx(w.PrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     w.pendingNonce,
		To:        &to,
		Value:     value,
		Gas:       gasUnits,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		return err
	}
	ttx := &TrackedTransaction{
		Transaction: tx,
		timeSent:    time.Now(),
	}

	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	w.pendingTransactions = append(w.pendingTransactions, ttx)
	return w.client.SendTransaction(context.Background(), tx)
}

func (w *Wallet) reSendTx(latestBlock *types.Block, oldTx *TrackedTransaction) error {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	gasUnits, baseFee, gasTipCapFloat, err := w.estimateGas(latestBlock, tx.Transaction.To())
	if err != nil {
		return err
	}
	gasTipCap := new(big.Int)
	gasTipCapFloat.Int(gasTipCap)
	gasFeeCap := big.NewInt(0).Add(baseFee, gasTipCap)

	newTx, err := types.SignNewTx(w.PrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     w.pendingNonce,
		To:        &to,
		Value:     value,
		Gas:       gasUnits,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		return err
	}
	tx = &TrackedTransaction{
		Transaction: newTx,
		timeSent:    time.Now(),
	}
}

// stopTracking stops tracking a transaction and removes it from the pending transactions list
func (w *Wallet) stopTracking(txHash common.Hash) {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	for index, tx := range w.pendingTransactions {
		if tx.Transaction.Hash() == txHash {
			log.Trace().Str("Hash", txHash.Hex()).Msg("Stopped tracking transaction")
			w.pendingTransactions = append(w.pendingTransactions[:index], w.pendingTransactions[index+1:]...)
			return
		}
	}
}

// estimateGas estimates the gas for the transaction, and the gas fee cap as a float for easier multiplication
func (w *Wallet) estimateGas(latestBlock *types.Block, to *common.Address) (
	gas uint64,
	baseFee *big.Int,
	gasTipCap *big.Float,
	err error,
) {
	msg := ethereum.CallMsg{
		From: w.Address,
		To:   to,
	}
	gas, err = w.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, nil, nil, err
	}
	suggestedGasTipCap, err := w.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return 0, nil, nil, err
	}
	gasTipCap = big.NewFloat(0).SetInt64(suggestedGasTipCap.Int64())
	baseFee = big.NewInt(0).Mul(latestBlock.BaseFee(), big.NewInt(2))
	return gas, baseFee, gasTipCap, err
}
