package fans

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

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

	client              *ethclient.Client
	pendingNonce        uint64
	pendingMu           sync.Mutex
	pendingTransactions []*TrackedTransaction
	crazedLevel         *config.CrazedLevel
}

// NewWallet creates a new wallet with a new private key
func NewWallet(client *ethclient.Client, crazedLevel *config.CrazedLevel) (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return LoadWallet(client, crazedLevel, privateKey)
}

// LoadWallet creates a new wallet with a supplied private key
func LoadWallet(
	client *ethclient.Client,
	crazedLevel *config.CrazedLevel,
	privateKey *ecdsa.PrivateKey,
) (*Wallet, error) {
	pendingNonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		return nil, err
	}
	return &Wallet{
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),

		client:              client,
		pendingNonce:        pendingNonce,
		pendingTransactions: []*TrackedTransaction{},
		crazedLevel:         crazedLevel,
	}, nil
}

// UpdatePendingTxs updates the wallet's pending transactions, removing mined transactions and resending timed out transactions
func (w *Wallet) UpdatePendingTxs(block *types.Block) {
	for index, tx := range w.pendingTransactions {
		if block.Transaction(tx.Transaction.Hash()) != nil { // transaction was mined
			w.pendingTransactions = append(w.pendingTransactions[:index], w.pendingTransactions[index+1:]...)
			continue
		}
		if time.Since(tx.timeSent) > w.crazedLevel.TransactionTimeout { // transaction timed out, resend
			w.pendingTransactions = append(w.pendingTransactions[:index], w.pendingTransactions[index+1:]...)
			oldTip := big.NewFloat(0).SetInt64(tx.Transaction.GasTipCap().Int64())
			newGasTip := oldTip.Mul(oldTip, big.NewFloat(w.crazedLevel.GasPriceMultiplier))
			newGasUint, _ := newGasTip.Uint64()
			w.sendTx(tx.Transaction.Nonce(), big.NewInt(0).SetUint64(newGasUint))
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
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	tx := &types.DynamicFeeTx{
		ChainID: config.Current.BigChainID,
		Nonce:   w.pendingNonce,
		To:      &to,
		Value:   value,
		Gas:     21_000,
	}
	unsignedTx.S
	ttx := &TrackedTransaction{
		Transaction: tx,
		timeSent:    time.Now(),
	}
	w.pendingTransactions = append(w.pendingTransactions, ttx)
	w.sendTx(latestBlock, nonce, gasTipCap)
}

func (w *Wallet) sendTx(latestBlock *types.Block, nonce uint64, gasTipCap *big.Int) error {
	baseFee := big.NewInt(0).Mul(latestBlock.BaseFee(), big.NewInt(2))
	gasFeeCap := big.NewInt(0).Add(baseFee, gasTipCap)

	if err != nil {
		log.Error().Err(err).
			Uint64("Nonce", nonce).
			Uint64("Gas Tip Cap", gasTipCap.Uint64()).
			Uint64("Gas Fee Cap", gasFeeCap.Uint64()).
			Msg("Error signing transaction")
		return err
	}
}

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
