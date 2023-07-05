package fans

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/convert"
)

type trackedTransaction struct {
	tx       *types.Transaction
	timeSent time.Time
}

var sendAmount = big.NewInt(42069)

// Fan is an NFT fan that will search for NFTs
type Fan struct {
	Address    *common.Address
	PrivateKey *ecdsa.PrivateKey
	Flutter    *big.Float

	balance             *big.Int
	pendingNonce        uint64
	trackedTransactions map[common.Hash]trackedTransaction
	trackedMu           sync.RWMutex
	client              *ethclient.Client
}

// New creates a new fan
func New(client *ethclient.Client) (*Fan, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	addr, err := convert.PrivateKeyToAddress(key)
	if err != nil {
		return nil, err
	}
	nonce, err := client.PendingNonceAt(context.Background(), *addr)
	if err != nil {
		return nil, err
	}

	return &Fan{
		Address:    addr,
		PrivateKey: key,
		Flutter:    big.NewFloat(1.1),

		balance:             big.NewInt(0),
		pendingNonce:        nonce,
		trackedTransactions: map[common.Hash]trackedTransaction{},
		client:              client,
	}, nil
}

// ReceiveBlock receives a new block from the chain, and updates pending transactions accordingly
func (f *Fan) ReceiveBlock(newBlock *types.Block) error {
	f.trackedMu.Lock()
	defer f.trackedMu.Unlock()
	for _, tx := range newBlock.Transactions() {
		if _, ok := f.trackedTransactions[tx.Hash()]; ok {
			delete(f.trackedTransactions, tx.Hash())
			log.Trace().Str("Hash", tx.Hash().Hex()).Msg("Confirmed transaction")
		}
	}
	_, err := f.SendRandomTransaction(newBlock.BaseFee())
	return err
}

func (f *Fan) SendRandomTransaction(baseFee *big.Int) (common.Hash, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		log.Error().Err(err).Msg("Error generating key")
		return common.Hash{}, err
	}
	addr, err := convert.PrivateKeyToAddress(key)
	if err != nil {
		log.Error().Err(err).Msg("Error generating address")
		return common.Hash{}, err
	}
	gasTipCap, gasFeeCap, err := f.calculateGas(baseFee)
	if err != nil {
		log.Error().Err(err).Msg("Error calculating gas")
		return common.Hash{}, err
	}
	if gasFeeCap.Cmp(f.balance) >= 0 {
		return common.Hash{}, fmt.Errorf("not enough balance to send transaction")
	}
	f.balance.Sub(f.balance, gasFeeCap)
	f.balance.Sub(f.balance, sendAmount)
	tx, err := types.SignNewTx(f.PrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     f.pendingNonce,
		To:        addr,
		Value:     sendAmount,
		Gas:       21_000,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error signing transaction")
		return common.Hash{}, err
	}
	f.pendingNonce++
	err = f.client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Error().Err(err).
			Str("Hash", tx.Hash().Hex()).
			Uint64("Gas Tip Cap", gasTipCap.Uint64()).
			Uint64("Base Fee", baseFee.Uint64()).
			Msg("Error sending transaction")
		return common.Hash{}, err
	}
	f.trackedTransactions[tx.Hash()] = trackedTransaction{
		tx:       tx,
		timeSent: time.Now(),
	}
	log.Trace().
		Str("Hash", tx.Hash().Hex()).
		Uint64("Gas Tip Cap", gasTipCap.Uint64()).
		Uint64("Base Fee", baseFee.Uint64()).
		Msg("Sent transaction")
	return tx.Hash(), nil
}

// gasTipCap = floorPrice + floor(Flutter * random, peak)
func (f *Fan) calculateGas(baseFee *big.Int) (gasTipCap, gasFeeCap *big.Int, err error) {
	limitGasBig := big.NewInt(0).Sub(config.Current.PeakGasPriceWei, config.Current.FloorGasPriceWei)
	random, err := rand.Int(rand.Reader, limitGasBig)
	if err != nil {
		return nil, nil, err
	}
	randFloat := big.NewFloat(0).SetInt(random)

	uncertaintyFloat := big.NewFloat(0).Mul(f.Flutter, randFloat)
	gasTipCap, _ = uncertaintyFloat.Int(nil)
	gasTipCap.Add(gasTipCap, config.Current.FloorGasPriceWei)

	gasFeeCap = big.NewInt(0).Add(baseFee, gasTipCap)
	return gasTipCap, gasFeeCap, nil
}

func (f *Fan) Fund(wei *big.Int, fundingNonce uint64, timeout time.Duration) error {
	latestHeader, err := f.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	baseFee := new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2))
	tipCap, err := f.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}
	gasFeeCap := big.NewInt(0).Add(baseFee, tipCap)
	tx, err := types.SignNewTx(config.Current.FundingPrivateKey, types.LatestSignerForChainID(config.Current.BigChainID), &types.DynamicFeeTx{
		ChainID:   config.Current.BigChainID,
		Nonce:     fundingNonce,
		To:        f.Address,
		Value:     wei,
		Gas:       21_000,
		GasTipCap: tipCap,
		GasFeeCap: gasFeeCap,
	})
	if err != nil {
		return err
	}

	err = f.client.SendTransaction(context.Background(), tx)
	if err != nil {
		return err
	}
	log.Trace().Str("Hash", tx.Hash().Hex()).Uint64("Nonce", fundingNonce).Uint64("Wei", wei.Uint64()).Msg("Funding fan")
	f.trackedTransactions[tx.Hash()] = trackedTransaction{
		tx:       tx,
		timeSent: time.Now(),
	}
	if err = f.ConfirmTransaction(tx.Hash(), timeout); err != nil {
		return err
	}
	f.balance.Add(f.balance, wei)
	return nil
}

func (f *Fan) ConfirmTransaction(txHash common.Hash, timeout time.Duration) error {
	f.trackedMu.RLock()
	if _, ok := f.trackedTransactions[txHash]; !ok {
		return fmt.Errorf("transaction %s not found", txHash.Hex())
	}
	f.trackedMu.RUnlock()

	timeoutC := time.After(timeout)
	check := time.NewTicker(500 * time.Millisecond)
	defer check.Stop()
	for {
		select {
		case <-timeoutC:
			return fmt.Errorf("error confirming tx %s after %s", txHash.Hex(), timeout)
		case <-check.C:
			f.trackedMu.RLock()
			_, ok := f.trackedTransactions[txHash]
			f.trackedMu.RUnlock()
			if !ok {
				log.Trace().Str("Hash", txHash.Hex()).Msg("Confirmed transaction")
				return nil
			}
		}
	}
}
