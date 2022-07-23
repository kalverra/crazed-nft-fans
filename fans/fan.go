// Package fans details the actions of each crazed fan
package fans

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
)

// Fan is an NFT fan that searches for the NFT by incessantly bumping gas
type Fan struct {
	ID          string
	Name        string
	Client      *client.EthClient
	PrivateKey  *ecdsa.PrivateKey
	Address     common.Address
	CrazedLevel int

	currentlySearching bool
	searchingMutex     sync.Mutex
	stopSearch         chan struct{}
}

// NewFan generates a new fan
func NewFan() (*Fan, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	address, err := client.PrivateKeyToAddress(privateKey)
	if err != nil {
		return nil, err
	}
	fanClient, err := client.NewClient()
	if err != nil {
		return nil, err
	}

	return &Fan{
		ID:          uuid.New().String(),
		Name:        newName(),
		Client:      fanClient,
		PrivateKey:  privateKey,
		Address:     address,
		CrazedLevel: config.Current.GetCrazedLevel(),

		currentlySearching: false,
		stopSearch:         make(chan struct{}),
	}, nil
}

// Search triggers the fan to search for the hidden NFT
func (f *Fan) Search() error {
	log.Info().Str("Fan", f.Name).Msg("Starting Search")
	f.searchingMutex.Lock()
	f.currentlySearching = true
	f.searchingMutex.Unlock()
	defer func() {
		log.Info().Str("Fan", f.Name).Msg("Stopping Search")
		f.searchingMutex.Lock()
		f.currentlySearching = false
		f.searchingMutex.Unlock()
	}()

	newHeads := make(chan *types.Header)
	sub, err := f.Client.SubscribeNewBlocks(context.Background(), newHeads)
	if err != nil {
		return err
	}

	for {
		select {
		case err = <-sub.Err():
			return err
		case <-f.stopSearch:
			return nil
		}
	}
}

// StopSearch halts the fan if it is currently searching for the NFT
func (f *Fan) StopSearch() {
	f.stopSearch <- struct{}{}
}

// IsSearching indicates if the fan is currently in the midst of searching for an NFT
func (f *Fan) IsSearching() bool {
	f.searchingMutex.Lock()
	defer f.searchingMutex.Unlock()
	return f.currentlySearching
}

// Fund funds the fan with the provided amount of ETH. Errors if the tx doesn't complete
func (f *Fan) Fund(ctx context.Context, amount *big.Float) error {
	hash, err := f.Client.SendTransaction(f.Client.FundedKey, f.Address, big.NewInt(0), amount)
	if err != nil {
		return err
	}
	confirmed, err := f.Client.ConfirmTransaction(ctx, hash)
	if err != nil {
		return err
	}
	if !confirmed {
		return fmt.Errorf("Unable to confirm funding tx for fan %s", f.Name)
	}
	return nil
}
