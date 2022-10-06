package fans

import (
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
)

type FanPresident struct {
	fans      []*Fan
	fansMutex sync.Mutex

	wallet     *client.Wallet
	privateKey *ecdsa.PrivateKey
	nonce      uint64
}

func NewPresident() (*FanPresident, error) {
	wallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	if err != nil {
		return nil, err
	}

	return &FanPresident{
		fans: make([]*Fan, 0),

		wallet:     wallet,
		privateKey: config.Current.FundingPrivateKey,
		nonce:      0,
	}, nil
}

func (fp *FanPresident) NewFans(number uint) error {
	fp.fansMutex.Lock()
	defer fp.fansMutex.Unlock()
	for i := 0; i < int(number); i++ {
		fan, err := NewFan()
		if err != nil {
			return err
		}
		fp.fans = append(fp.fans, fan)
	}
	return nil
}

func (fp *FanPresident) FundFans(etherAmount *big.Float) {
	var wg sync.WaitGroup
	for _, f := range fp.fans {
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

func (fp *FanPresident) Fans() []*Fan {
	fp.fansMutex.Lock()
	defer fp.fansMutex.Unlock()
	return fp.fans
}
