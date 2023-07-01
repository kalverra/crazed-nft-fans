package fans

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

type President struct {
	Client            *ethclient.Client
	LatestBlock       *types.Block
	LatestBlockNumber uint64
	PrivateKey        *ecdsa.PrivateKey
	blockUpdateMu     sync.Mutex

	wallet             *Wallet
	stopSearchingBlock uint64
	stopSearchingTime  time.Time
	stopSearchingMu    sync.Mutex
	fans               []*Fan
	fansMu             sync.Mutex
}

// NewPresident creates a new president and starts watching the chain
func NewPresident() (*President, error) {
	client, err := ethclient.Dial(config.Current.WS)
	if err != nil {
		return nil, err
	}
	pres := &President{
		fans:       []*Fan{},
		Client:     client,
		PrivateKey: config.Current.FundingPrivateKey,
	}

	wallet, err := LoadWallet(config.Current.FundingPrivateKey, fmt.Sprint(rand.Uint64()), "president", pres.Client, config.President)
	if err != nil {
		return nil, err
	}
	pres.wallet = wallet
	return pres, pres.watch()
}

func NewPresidentWithWallet(wallet *Wallet) (*President, error) {
	client, err := ethclient.Dial(config.Current.WS)
	if err != nil {
		return nil, err
	}
	pres := &President{
		wallet:     wallet,
		fans:       []*Fan{},
		Client:     client,
		PrivateKey: config.Current.FundingPrivateKey,
	}
	return pres, pres.watch()
}

// RecruitFans recruits a number of new fans to the president's cause
func (p *President) RecruitFans(count int) error {
	log.Info().Int("Fan Count", count).Msg("Recruiting fans")
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for i := 0; i < count; i++ {
		fan, err := New(p.Client, config.Current.GetCrazedLevel())
		if err != nil {
			return err
		}
		p.fans = append(p.fans, fan)
	}
	return nil
}

// FundFans funds all the president's fans
func (p *President) FundFans(value *big.Int) error {
	log.Info().Uint64("Value", value.Uint64()).Msg("Funding Fans")
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for _, fan := range p.fans {
		_, err := p.wallet.SendTransaction(p.LatestBlock, fan.Address, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ActivateFans activates all fans to start searching
func (p *President) ActivateFans() {
	log.Info().Msg("Activating fans")
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for _, fan := range p.fans {
		fan.Search()
	}
}

// ActivateFansTimeSpan activates all fans to start searching for a given duration, returning at the end of that duration
func (p *President) ActivateFansTimeSpan(dur time.Duration) {
	log.Info().Str("Duration", dur.String()).Msg("Activating fans for duration")
	p.stopSearchingMu.Lock()
	defer p.stopSearchingMu.Unlock()
	p.stopSearchingTime = time.Now().Add(dur)
	p.ActivateFans()
}

// ActivateFansBlockSpan activates all fans to start searching for a given number of blocks, returning at the end of that duration
func (p *President) ActivateFansBlockSpan(blocks int) {
	log.Info().Int("Blocks", blocks).Msg("Activating fans for block span")
	p.stopSearchingMu.Lock()
	defer p.stopSearchingMu.Unlock()
	p.stopSearchingBlock = p.GetLatestBlockNumber() + uint64(blocks)
	p.ActivateFans()
}

// StopFans stops all fans from searching
func (p *President) StopFans() {
	log.Info().Msg("Stopping fans")
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	for _, fan := range p.fans {
		fan.Stop()
	}
}

// Fans returns all current fans
func (p *President) Fans() []*Fan {
	p.fansMu.Lock()
	defer p.fansMu.Unlock()
	return p.fans
}

// watch watches the chain, informing fans of new blocks
func (p *President) watch() error {
	headerChannel := make(chan *types.Header)
	sub, err := p.Client.SubscribeNewHead(context.Background(), headerChannel)
	if err != nil {
		return err
	}
	header, err := p.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	p.LatestBlock, err = p.Client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	p.LatestBlockNumber = header.Number.Uint64()

	go func() {
		defer sub.Unsubscribe()
		for { // We're assuming that the RPC/chain is a reliable one. This is a false assumption if ever running in the real world
			select {
			case err = <-sub.Err():
				sub.Unsubscribe()
				log.Error().Err(err).Msg("Error in subscription, retrying")

				sub, err = p.Client.SubscribeNewHead(context.Background(), headerChannel)
				if err != nil {
					log.Error().Err(err).Msg("Error in subscription, retrying")
				}
			case header := <-headerChannel:
				var block *types.Block
				log.Debug().
					Str("Hash", header.Hash().Hex()).
					Uint64("Number", header.Number.Uint64()).
					Uint64("Gas Used", header.GasUsed).
					Uint64("Gas Limit", header.GasLimit).
					Msg("New Header")
					// Get the full block so we can comb through the transactions ourselves and not overload the chain/RPC
				block, err = p.Client.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					log.Error().Err(err).Str("Hash", header.Hash().Hex()).Uint64("Number", header.Number.Uint64()).Msg("Error getting block from header")
					continue
				}
				p.blockUpdateMu.Lock()
				p.LatestBlock = block
				p.LatestBlockNumber++
				p.blockUpdateMu.Unlock()

				_, err = p.wallet.UpdatePendingTxs(block)
				if err != nil {
					log.Error().Err(err).
						Str("Block Hash", block.Hash().Hex()).
						Uint64("Block Number", block.NumberU64()).
						Msg("Error updating pending transactions for president")
				}
				p.fansMu.Lock()
				for _, fan := range p.fans {
					go fan.ReceiveBlock(block)
				}
				p.fansMu.Unlock()

				// Check if we need to stop searching after some prescribed time or block range
				p.stopSearchingMu.Lock()
				if p.stopSearchingBlock > 0 && p.GetLatestBlockNumber() >= p.stopSearchingBlock {
					p.StopFans()
					log.Info().Uint64("Stop Block", p.stopSearchingBlock).Msg("Stopped fan search")
					p.stopSearchingBlock = 0
				}
				if !p.stopSearchingTime.IsZero() && time.Now().After(p.stopSearchingTime) {
					p.StopFans()
					log.Info().Time("Stop Time", p.stopSearchingTime).Msg("Stopped fan search")
					p.stopSearchingTime = time.Time{}
				}
				p.stopSearchingMu.Unlock()
			}
		}
	}()
	return nil
}

func (p *President) GetLatestBlock() *types.Block {
	p.blockUpdateMu.Lock()
	defer p.blockUpdateMu.Unlock()
	return p.LatestBlock
}

func (p *President) GetLatestBlockNumber() uint64 {
	p.blockUpdateMu.Lock()
	defer p.blockUpdateMu.Unlock()
	return p.LatestBlockNumber
}
