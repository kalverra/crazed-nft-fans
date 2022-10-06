package fans

import (
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
)

type FanPresident struct {
	Fans []*Fan

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
		Fans: make([]*Fan, 0),

		wallet:     wallet,
		privateKey: config.Current.FundingPrivateKey,
		nonce:      0,
	}, nil
}

func (fp *FanPresident) NewFans(number uint) error {
	for i := 0; i < int(number); i++ {
		fan, err := NewFan()
		if err != nil {
			return err
		}
		fp.Fans = append(fp.Fans, fan)
	}
	return nil
}

func (fp *FanPresident) FundFans(etherAmount *big.Float) {
	var wg sync.WaitGroup
	for _, f := range fp.Fans {
		fan := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			fp.wallet.SendTransaction(time.Minute, big.NewFloat(1.125), fan.Address, big.NewInt(1), etherAmount)
		}()
	}
	wg.Wait()
}

func (fp *FanPresident) BalanceAt() {

}
