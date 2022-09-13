package client

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

var (
	GlobalTransactionTracker *transactionTracker
	alreadyInitiated         = false
)

func NewTransactionTracker() error {
	if alreadyInitiated {
		return fmt.Errorf("cannot initialize global tracker more than once; only once tracker allowed")
	}
	alreadyInitiated = true

	trackerClient, err := NewClient(config.Current.WS)
	if err != nil {
		return err
	}
	tracker := map[common.Address]*trackedAddress{}

	log.Info().Msg("Instantiated Global Transaction Tracker")
	GlobalTransactionTracker = &transactionTracker{
		transactions: tracker,
		mutex:        &sync.Mutex{},
		client:       trackerClient,
	}
	return nil
}

type transactionTracker struct {
	transactions map[common.Address]*trackedAddress // fromAddress: addressInfo
	mutex        *sync.Mutex
	client       *EthClient
}

type trackedAddress struct {
	lastUsedNonce       uint64
	trackedTransactions map[common.Hash]*trackedTransaction
}

// trackedTransaction has all details of the transaction
type trackedTransaction struct {
	to        common.Address
	gasFeeCap *big.Int
	gasTipCap *big.Int
	nonce     uint64
	status    txStatus
}

// txStatus possible status for a transaction
type txStatus string

// all viable transaction statuses
const (
	pending    txStatus = "Pending"
	successful txStatus = "Successful"
	abandoned  txStatus = "Abandoned"
)

// NewTransaction starts tracking a transaction
func (t *transactionTracker) NewTransaction(from common.Address, newTx *types.Transaction) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	oldTx := t.getTransactionByNonce(from, newTx.Nonce())
	if oldTx != nil { // Replace an old Transaction
		oldTx.status = abandoned
	}
	t.transactions[from].trackedTransactions[newTx.Hash()] = &trackedTransaction{
		to:        *newTx.To(),
		gasFeeCap: newTx.GasFeeCap(),
		gasTipCap: newTx.GasTipCap(),
		nonce:     newTx.Nonce(),
		status:    pending,
	}
	log.Debug().Str("From", from.Hex()).
		Str("Hash", newTx.Hash().Hex()).
		Uint64("Nonce", newTx.Nonce()).
		Msg("Tracking New Transaction")
}

// CompletedTransaction indicates that a transaction has been confirmed
func (t *transactionTracker) CompletedTransaction(from common.Address, hash common.Hash) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	addressTxs, isTracked := t.transactions[from]
	if !isTracked {
		return fmt.Errorf("address '%s' is not tracked, but attempted to mark as complete", from.Hex())
	}
	trackedTx, isTracked := addressTxs.trackedTransactions[hash]
	if !isTracked {
		return fmt.Errorf("tx '%s' from '%s' is not tracked, but attempted to mark as complete", from.Hex(), hash.Hex())
	}
	trackedTx.status = successful
	log.Debug().Str("From", from.Hex()).Str("Hash", hash.Hex()).Uint64("Nonce", trackedTx.nonce).Msg("Completed Transaction")
	return nil
}

// FirstAvailableNonce retrieves the first nonce, to our knowledge, that has not been used yet in either completed OR
// pending txs.
func (t *transactionTracker) FirstAvailableNonce(from common.Address) uint64 {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var (
		err          error
		pendingNonce uint64
	)
	_, isTracked := t.transactions[from]
	if !isTracked { // If we haven't tracked that address, we need to see its state on chain
		pendingNonce, err = t.client.innerClient.PendingNonceAt(context.Background(), from)
		if err != nil {
			log.Fatal().Err(err).Msg("Error retrieving nonce")
		}
		t.transactions[from] = &trackedAddress{
			lastUsedNonce:       pendingNonce,
			trackedTransactions: map[common.Hash]*trackedTransaction{},
		}
		return pendingNonce
	}
	t.transactions[from].lastUsedNonce++
	return t.transactions[from].lastUsedNonce
}

// FirstAvailableNonce retrieves the first nonce, to our knowledge, that has not been used yet in either completed OR
// pending txs.
func (t *transactionTracker) FundingNonce() uint64 {
	fundingAddr, err := PrivateKeyToAddress(config.Current.FundingPrivateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Error converting key to address")
	}
	return t.FirstAvailableNonce(fundingAddr)
}

// getTransactionByNonce retrieves a transaction with the supplied nonce, nil if there is none
func (t *transactionTracker) getTransactionByNonce(from common.Address, nonce uint64) *trackedTransaction {
	for _, tx := range t.transactions[from].trackedTransactions {
		if tx.nonce == nonce {
			return tx
		}
	}
	return nil
}
