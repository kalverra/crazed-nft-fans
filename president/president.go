package president

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
)

type TrackedBlock struct {
	Hash     string `json:"hash"`
	Number   uint64 `json:"number"`
	GasPrice uint64 `json:"gasPrice"`
	BaseFee  uint64 `json:"baseFee"`
}

var (
	trackedMu      sync.RWMutex
	trackedBlocks  = map[uint64]*TrackedBlock{}
	fanClub        = []*fans.Fan{}
	client         *ethclient.Client
	fundingNonceMu sync.Mutex
	fundingNonce   uint64
)

func WatchChain() error {
	var err error
	client, err = ethclient.Dial(config.Current.WS)
	if err != nil {
		return err
	}

	fundingNonce, err = client.PendingNonceAt(context.Background(), config.Current.FundingAddress)
	if err != nil {
		return err
	}

	newHeaderChan := make(chan *types.Header)
	subscription, err := client.SubscribeNewHead(context.Background(), newHeaderChan)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case err := <-subscription.Err():
				log.Error().Err(err).Msg("Error in subscription")

				subscription, err = client.SubscribeNewHead(context.Background(), newHeaderChan)
				if err != nil {
					log.Error().Err(err).Msg("Error re-subscribing")
				}
			case header := <-newHeaderChan:
				gasPrice, err := client.SuggestGasPrice(context.Background())
				if err != nil {
					log.Error().Err(err).Uint64("Header", header.Number.Uint64()).Msg("Error getting gas price")
					continue
				}
				log.Info().
					Str("Hash", header.Hash().Hex()).
					Uint64("Number", header.Number.Uint64()).
					Uint64("Gas Price", gasPrice.Uint64()).
					Uint64("Base Fee", header.BaseFee.Uint64()).
					Uint64("Gas Limit", header.GasLimit).
					Uint64("Gas Used", header.GasUsed).
					Msg("New block")
				trackedBlock := &TrackedBlock{
					Hash:     header.Hash().String(),
					Number:   header.Number.Uint64(),
					GasPrice: gasPrice.Uint64(),
					BaseFee:  header.BaseFee.Uint64(),
				}
				TrackBlock(trackedBlock)
				block, err := client.BlockByNumber(context.Background(), header.Number)
				if err != nil {
					log.Error().Err(err).Uint64("Header", header.Number.Uint64()).Msg("Error getting block")
					continue
				}
				eg := errgroup.Group{}
				for _, f := range fanClub {
					fan := f
					eg.Go(func() error {
						return fan.ReceiveBlock(block)
					})
				}
				if err = eg.Wait(); err != nil {
					log.Error().Err(err).Uint64("Header", header.Number.Uint64()).Msg("Error receiving block")
				}
			}
		}
	}()

	return nil
}

func AllBlocks() []*TrackedBlock {
	trackedMu.RLock()
	defer trackedMu.RUnlock()

	blocks := []*TrackedBlock{}
	validBlockFlag := false
	for blockNum := 0; ; blockNum++ { // TODO: Save state of first block so we don't have to start from 0
		if block, ok := trackedBlocks[uint64(blockNum)]; !ok {
			if validBlockFlag { // end of valid blocks
				break
			}
		} else {
			validBlockFlag = true
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func BlocksSinceNumber(number uint64) []*TrackedBlock {
	trackedMu.RLock()
	defer trackedMu.RUnlock()
	number++
	blocks := []*TrackedBlock{}
	for block, ok := trackedBlocks[number]; ok; number++ {
		blocks = append(blocks, block)
	}
	return blocks
}

// TrackBlock adds another block to our tracked group
func TrackBlock(block *TrackedBlock) {
	trackedMu.Lock()
	defer trackedMu.Unlock()

	trackedBlocks[block.Number] = block
}

func FundFans(wei *big.Int) error {
	log.Info().Str("Wei", wei.String()).Int("Count", len(fanClub)).Msg("Funding fans")
	eg := errgroup.Group{}
	for _, f := range fanClub {
		fan := f
		eg.Go(func() error {
			return fan.Fund(wei, FundingNonce(), time.Minute)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	log.Info().Str("Wei", wei.String()).Int("Count", len(fanClub)).Msg("Funded fans")
	return nil
}

func RecruitFans(count int) error {
	for i := 0; i < count; i++ {
		fan, err := fans.New(client)
		if err != nil {
			return err
		}
		fanClub = append(fanClub, fan)
	}
	log.Info().Int("Count", count).Msg("Recruited fans")
	return nil
}

func FundingNonce() uint64 {
	fundingNonceMu.Lock()
	defer fundingNonceMu.Unlock()
	n := fundingNonce
	fundingNonce++
	log.Trace().Uint64("Old", n).Uint64("New", fundingNonce).Msg("Funding nonce incremented")
	return n
}
