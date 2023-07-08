package president

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/convert"
	"github.com/kalverra/crazed-nft-fans/fans"
)

type TrackedBlock struct {
	Hash     string `json:"hash"`
	Number   uint64 `json:"number"`
	GasPrice uint64 `json:"gasPrice"`
	BaseFee  uint64 `json:"baseFee"`
}

var (
	fanClub = []*fans.Fan{}
	client  *ethclient.Client

	trackedMu     sync.RWMutex
	trackedBlocks = map[uint64]*TrackedBlock{}

	fundingNonceMu sync.Mutex
	fundingNonce   uint64

	TargetGasPrice    = big.NewInt(35000000000) // 35 gwei, a common baseline
	gasPriceIncrement = big.NewInt(1000000000)  // 1 gwei
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
				percentBlockFilled := (float64(header.GasUsed) / float64(header.GasLimit)) * 100
				log.Info().
					Str("Hash", header.Hash().Hex()).
					Uint64("Number", header.Number.Uint64()).
					Uint64("Gas Price", gasPrice.Uint64()).
					Uint64("Base Fee", header.BaseFee.Uint64()).
					Uint64("Gas Limit", header.GasLimit).
					Uint64("Gas Used", header.GasUsed).
					Str("Percent Block Filled", fmt.Sprintf("%.2f%%", percentBlockFilled)).
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
						return fan.ReceiveBlock(block, TargetGasPrice)
					})
				}
				if err = eg.Wait(); err != nil {
					if strings.Contains(err.Error(), "insufficient funds") {
						log.Warn().Msg("Fans out of money, deploying capital!")
						go func() {
							err = FundFans(convert.EtherToWei(big.NewFloat(100)))
							if err != nil {
								log.Error().Err(err).Msg("Error funding fans, app is in a bad state")
							}
						}()
					} else {
						log.Error().Err(err).Uint64("Header", header.Number.Uint64()).Msg("Error receiving block")
					}
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
	return n
}

func SetGasTarget(gasPrice *big.Int) {
	TargetGasPrice = gasPrice
}

func IncreaseGasTarget() *big.Int {
	newLevel := big.NewInt(0).Add(TargetGasPrice, gasPriceIncrement)
	SetGasTarget(newLevel)
	return newLevel
}

func DecreaseGasTarget() *big.Int {
	newLevel := big.NewInt(0).Sub(TargetGasPrice, gasPriceIncrement)
	SetGasTarget(newLevel)
	return newLevel
}

func Spike() *big.Int {
	log.Info().Msg("Spiking Gas Price")
	newLevel := big.NewInt(0).Mul(TargetGasPrice, big.NewInt(100))
	SetGasTarget(newLevel)
	return newLevel
}
