// Package fans details the actions of each crazed fan
package fans

import (
	"context"
	"crypto/ecdsa"
	"fmt"
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
	"github.com/kalverra/crazed-nft-fans/guzzle"
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
				f.searchTx()
			}
		}
	}()
	return nil
}

func (f *Fan) searchTx() {
	privateKey, err := crypto.GenerateKey()
	if err != nil { // Error handling is hard
		log.Fatal().Err(err).Msg("Error generating crypto key")
	}
	address, err := client.PrivateKeyToAddress(privateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Error converting private key to address")
	}
	reSendTime, reSendMult := f.reSendVals()
	go f.Wallet.SendTransaction(
		reSendTime,
		reSendMult,
		address,
		big.NewInt(1),
		client.WeiToEther(big.NewInt(1)),
	)
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

func (f *Fan) reSendVals() (reSendTime time.Duration, reSendGasMultiple *big.Float) {
	switch f.CrazedLevel {
	case 1:
		reSendTime = time.Minute * 2
		reSendGasMultiple = big.NewFloat(1.125)
	case 2:
		reSendTime = time.Minute
		reSendGasMultiple = big.NewFloat(1.15)
	case 3:
		reSendTime = time.Second * 45
		reSendGasMultiple = big.NewFloat(1.3)
	case 4:
		reSendTime = time.Second * 30
		reSendGasMultiple = big.NewFloat(1.5)
	case 5:
		reSendTime = time.Second * 10
		reSendGasMultiple = big.NewFloat(2)
	}
	return
}

func (f *Fan) Guzzle(guzzler *guzzle.Guzzle, amountOfGas uint64) error {
	if amountOfGas >= 30_000_000 {
		return fmt.Errorf("%d of gas is larger than the block gas limit of 30,000,000", amountOfGas)
	}
	opts, err := f.Wallet.TransactionOpts()
	if err != nil {
		return err
	}
	log.Debug().Uint64("Amount", amountOfGas).Msg("Guzzling Gas")
	// opts.GasLimit = amountOfGas + 1
	tx, err := guzzler.Guzzle(opts, big.NewInt(0).SetUint64(amountOfGas))
	if err != nil {
		return err
	}
	return f.Wallet.ConfirmTxWait(context.Background(), tx)
}
