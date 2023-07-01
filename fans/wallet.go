package fans

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
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
	ID         string
	Name       string

	client              *ethclient.Client
	pendingNonce        uint64
	pendingMu           sync.Mutex
	pendingTransactions []*TrackedTransaction
	crazedLevel         *config.CrazedLevel
}

// NewFanWallet creates a new wallet designed to be used by a fan
func NewWallet(id, name string, client *ethclient.Client, crazedLevel *config.CrazedLevel) (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return LoadWallet(privateKey, id, name, client, crazedLevel)
}

// LoadWallet generates a wallet with an existing private key
func LoadWallet(privateKey *ecdsa.PrivateKey, id, name string, client *ethclient.Client, crazedLevel *config.CrazedLevel) (*Wallet, error) {
	pendingNonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		return nil, err
	}
	return &Wallet{
		ID:         id,
		Name:       name,
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),

		client:              client,
		pendingNonce:        pendingNonce,
		pendingTransactions: []*TrackedTransaction{},
		crazedLevel:         crazedLevel,
	}, nil
}

// UpdatePendingTxs updates the wallet's pending transactions, removing mined transactions and resending timed out transactions
// returns the number of transactions that are still pending
func (w *Wallet) UpdatePendingTxs(latestBlock *types.Block) (int, error) {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	log.Trace().
		Str("Name", w.Name).
		Str("ID", w.ID).
		Int("Pending Transactions", len(w.pendingTransactions)).
		Msg("Updating pending transactions")

	pendingTxs := []*TrackedTransaction{}
	for _, tx := range w.pendingTransactions {
		if latestBlock.Transaction(tx.Transaction.Hash()) == nil { // transaction not mined, still pending
			pendingTxs = append(pendingTxs, tx)
			continue
		}
		if time.Since(tx.timeSent) > w.crazedLevel.TransactionTimeout { // transaction timed out, resend
			if err := w.reSendTx(latestBlock, tx); err != nil {
				return 0, err
			}
		}
	}
	w.pendingTransactions = pendingTxs
	return len(w.pendingTransactions), nil
}

// PendingTransactions returns the wallet's currently pending transactions
func (w *Wallet) PendingTransactions() []*TrackedTransaction {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	return w.pendingTransactions
}

// SendTransaction sends a transaction to the chain and tracks and re-sends it if it times out. Returns the transaction hash
func (w *Wallet) SendTransaction(latestBlock *types.Block, to *common.Address, value *big.Int) (common.Hash, error) {
	gasUnits, baseFee, gasTipCapFloat, err := w.estimateGas(latestBlock, to)
	if err != nil {
		return common.Hash{}, err
	}
	gasTipCap := new(big.Int)
	gasTipCapFloat.Int(gasTipCap)
	gasFeeCap := big.NewInt(0).Add(baseFee, gasTipCap)

	tx, err := types.SignNewTx(w.PrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     w.pendingNonce,
		To:        to,
		Value:     value,
		Gas:       gasUnits,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		return common.Hash{}, err
	}
	ttx := &TrackedTransaction{
		Transaction: tx,
		timeSent:    time.Now(),
	}

	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()
	w.pendingNonce++
	w.pendingTransactions = append(w.pendingTransactions, ttx)

	log.Debug().
		Str("Name", w.Name).
		Str("ID", w.ID).
		Str("Hash", tx.Hash().Hex()).
		Uint64("Value", value.Uint64()).
		Str("To", to.Hex()).
		Msg("Sending new transaction")
	return tx.Hash(), w.client.SendTransaction(context.Background(), tx)
}

func (w *Wallet) reSendTx(latestBlock *types.Block, oldTx *TrackedTransaction) error {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	gasUnits, baseFee, gasTipCapFloat, err := w.estimateGas(latestBlock, oldTx.Transaction.To())
	if err != nil {
		return err
	}
	gasTipCapFloat.Mul(gasTipCapFloat, big.NewFloat(w.crazedLevel.GasPriceMultiplier))
	gasTipCap := new(big.Int)
	gasTipCapFloat.Int(gasTipCap)
	gasFeeCap := big.NewInt(0).Add(baseFee, gasTipCap)

	newTx, err := types.SignNewTx(w.PrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     oldTx.Transaction.Nonce(),
		To:        oldTx.Transaction.To(),
		Value:     oldTx.Transaction.Value(),
		Gas:       gasUnits,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		return err
	}
	log.Debug().
		Str("Old Hash", oldTx.Transaction.Hash().Hex()).
		Str("New Hash", newTx.Hash().Hex()).
		Str("Timeout", oldTx.timeSent.String()).
		Uint64("Old GasTipCap", oldTx.Transaction.GasTipCap().Uint64()).
		Uint64("New GasTipCap", newTx.GasTipCap().Uint64()).
		Msg("Re-sending transaction after timeout")
	oldTx = &TrackedTransaction{
		Transaction: newTx,
		timeSent:    time.Now(),
	}
	return w.client.SendTransaction(context.Background(), newTx)
}

// estimateGas estimates the gas for the transaction, and the gas fee cap as a float for easier multiplication
func (w *Wallet) estimateGas(latestBlock *types.Block, to *common.Address) (
	gas uint64,
	baseFee *big.Int,
	gasTipCap *big.Float,
	err error,
) {
	if latestBlock == nil {
		return 0, nil, nil, errors.New("latest block is nil")
	}
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
