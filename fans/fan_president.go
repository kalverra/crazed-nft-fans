package fans

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

type President struct {
	Fans   []*Fan
	Client *ethclient.Client
	fansMu sync.Mutex
}

// NewPresident creates a new president and starts watching the chain
func NewPresident() (*President, error) {
	client, err := ethclient.Dial(config.Current.WS)
	if err != nil {
		return nil, err
	}
	pres := &President{
		Fans:   []*Fan{},
		Client: client,
	}
	return pres, pres.watch()
}

// RecruitFans recruits a number of new fans to the president's cause
func (p *President) RecruitFans(count int) error {
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for i := 0; i < count; i++ {
		fan, err := New(config.Current.GetCrazedLevel())
		if err != nil {
			return err
		}
		p.Fans = append(p.Fans, fan)
	}
	return nil
}

// ActivateFans activates all fans to start searching
func (p *President) ActivateFans() {
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for _, fan := range p.Fans {
		fan.Search()
	}
}

// StopFans stops all fans from searching
func (p *President) StopFans() {
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for _, fan := range p.Fans {
		fan.Stop()
	}
}

// watch watches the chain, informing fans of new blocks
func (p *President) watch() error {
	headerChannel := make(chan *types.Header)
	sub, err := p.Client.SubscribeNewHead(context.Background(), headerChannel)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err = <-sub.Err():
				sub.Unsubscribe()
				log.Error().Err(err).Msg("Error in subscription, retrying")

				sub, err = p.Client.SubscribeNewHead(context.Background(), headerChannel)
				if err != nil {
					log.Error().Err(err).Msg("Error in subscription, retrying")
				}
			case header := <-headerChannel:
				log.Debug().
					Str("Hash", header.Hash().Hex()).
					Uint64("Number", header.Number.Uint64()).
					Uint64("Gas Used", header.GasUsed).
					Uint64("Gas Limit", header.GasLimit).
					Msg("New Header")
				p.fansMu.Lock()
				for _, fan := range p.Fans {
					go fan.ReceiveHeader(header)
				}
				p.fansMu.Unlock()
			}
		}
	}()
	return err
}
