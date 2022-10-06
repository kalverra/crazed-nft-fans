// Package fans details the actions of each crazed fan
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
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
)

// Fan is an NFT fan that searches for the NFT by incessantly bumping gas
type Fan struct {
	ID          string
	Name        string
	Wallet      *client.Wallet
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
	fanWallet, err := client.NewWallet(privateKey, config.Current.WS)
	if err != nil {
		return nil, err
	}

	return &Fan{
		ID:          uuid.New().String(),
		Name:        newName(),
		Wallet:      fanWallet,
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

	newHeads := make(chan *types.Header)
	sub, err := f.Wallet.SubscribeNewBlocks(context.Background(), newHeads)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err = <-sub.Err():
				log.Fatal().Str("Fan", f.Name).Err(err).Msg("Error while searching")
			case <-f.stopSearch:
				log.Info().Str("Fan", f.Name).Msg("Stopping Search")
				f.searchingMutex.Lock()
				f.currentlySearching = false
				f.searchingMutex.Unlock()
				return
			default:
				f.search()
			}
		}
	}()
	return nil
}

func (f *Fan) search() {
	to := common.HexToAddress("0x024c0763F8b55972Cd4a0349c79833bA9e3B2279")
	go f.Wallet.SendTransaction(time.Second*5, big.NewFloat(1.15), to, big.NewInt(1), big.NewFloat(.001))
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
