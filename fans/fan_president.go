package fans

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/guzzle"
)

var created = false

// FanPresident controls the meta behaviors of fans, creating new ones, and directing their activity
type FanPresident struct {
	fans      []*Fan
	fansMutex sync.Mutex

	gasGuzzlers []*guzzle.Guzzle
	guzzleMutex sync.Mutex

	wallet     *client.Wallet
	privateKey *ecdsa.PrivateKey
	address    common.Address
	nonce      uint64
}

// NewPresident creates a new president. Only can be used once
func NewPresident() (*FanPresident, error) {
	if created {
		return nil, errors.New("fan president already exists, there can only be one")
	}
	wallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	if err != nil {
		return nil, err
	}
	addr, err := client.PrivateKeyToAddress(config.Current.FundingPrivateKey)
	if err != nil {
		return nil, err
	}

	created = true
	return &FanPresident{
		fans: make([]*Fan, 0),

		wallet:     wallet,
		privateKey: config.Current.FundingPrivateKey,
		address:    addr,
		nonce:      0,
	}, nil
}

// NewFans generates a specified number of new fans
func (fp *FanPresident) NewFans(count int) error {
	fp.fansMutex.Lock()
	defer fp.fansMutex.Unlock()
	for i := 0; i < count; i++ {
		fan, err := NewFan()
		if err != nil {
			return err
		}
		fp.fans = append(fp.fans, fan)
	}
	return nil
}

// FundFans sends a specified amount of ether to all current fans
func (fp *FanPresident) FundFans(etherAmount *big.Float) {
	var wg sync.WaitGroup
	for _, f := range fp.Fans() {
		fan := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			floatFunding, _ := etherAmount.Float64()
			log.Info().
				Str("Address", fan.Address.Hex()).
				Str("Fan", fan.Name).
				Float64("Amount", floatFunding).
				Msg("Funding Fan")
			fp.wallet.SendTransaction(time.Minute, big.NewFloat(1.125), fan.Address, big.NewInt(1), etherAmount)
		}()
	}
	wg.Wait()
}

// Fans retrieves all the current fans
func (fp *FanPresident) Fans() []*Fan {
	fp.fansMutex.Lock()
	defer fp.fansMutex.Unlock()
	return fp.fans
}

// GasGuzzlers retrieves all active gas guzzling contracts
func (fp *FanPresident) GasGuzzlers() []*guzzle.Guzzle {
	fp.guzzleMutex.Lock()
	defer fp.guzzleMutex.Unlock()
	return fp.gasGuzzlers
}

// DeployGasGuzzlers deploys a specified number of gas guzzle contracts, waiting for them all to get on chain to move on
func (fp *FanPresident) DeployGasGuzzlers(count int) error {
	waitGroup := errgroup.Group{}
	for i := 0; i < count; i++ {
		waitGroup.Go(func() error {
			opts, err := fp.wallet.TransactionOpts()
			if err != nil {
				return err
			}
			_, tx, guzzler, err := guzzle.DeployGuzzle(opts, fp.wallet.EthClient)
			if err != nil {
				return err
			}
			err = fp.wallet.ConfirmTxWait(context.Background(), tx)
			if err != nil {
				return err
			}
			fp.guzzleMutex.Lock()
			defer fp.guzzleMutex.Unlock()
			fp.gasGuzzlers = append(fp.gasGuzzlers, guzzler)
			return nil
		})
	}
	return waitGroup.Wait()
}
