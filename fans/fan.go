package fans

import (
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/convert"
)

// Fan is an NFT fan that will search for NFTs
type Fan struct {
	ID          string
	Name        string
	Address     *common.Address
	PrivateKey  *ecdsa.PrivateKey
	CrazedLevel *config.CrazedLevel

	wallet             *Wallet
	client             *ethclient.Client
	stopButton         chan struct{}
	currentlySearching bool
	searchingMu        sync.Mutex
	blockChan          chan *types.Block
}

// New creates a new fan at a supplied crazed level
func New(client *ethclient.Client, level *config.CrazedLevel) (*Fan, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	fan := &Fan{
		ID:          xid.New().String(),
		Name:        generateName(),
		Address:     &addr,
		PrivateKey:  privateKey,
		CrazedLevel: level,

		client:             client,
		stopButton:         make(chan struct{}),
		currentlySearching: false,
		blockChan:          make(chan *types.Block),
	}
	fan.wallet, err = NewWallet(fan.ID, fan.Name, client, level)
	if err != nil {
		return nil, err
	}
	return fan, nil
}

// Search begins the fan's search
func (f *Fan) Search() {
	f.searchingMu.Lock()
	defer f.searchingMu.Unlock()
	f.currentlySearching = true
	log.Info().Str("ID", f.ID).Str("Name", f.Name).Msg("Fan Searching!")
	go f.searchLoop()
}

// searchLoop is the main loop for the fan's search
func (f *Fan) searchLoop() {
	for {
		select {
		case <-f.stopButton:
			f.searchingMu.Lock()
			log.Info().Str("ID", f.ID).Str("Name", f.Name).Msg("Stopping fan")
			f.currentlySearching = false
			f.searchingMu.Unlock()
			return
		case newBlock := <-f.blockChan:
			stillPendingCount, err := f.wallet.UpdatePendingTxs(newBlock)
			if err != nil {
				log.Error().Err(err).Str("Fan", f.Name).Str("ID", f.ID).Msg("Error updating pending transactions")
				continue
			}
			if stillPendingCount < f.CrazedLevel.MaxPendingTransactions {
				randomKey, err := crypto.GenerateKey()
				if err != nil {
					log.Error().Err(err).Msg("Error generating new private key")
					continue
				}
				randomAddr, err := convert.PrivateKeyToAddress(randomKey)
				if err != nil {
					log.Error().Err(err).Msg("Error generating new address")
					continue
				}
				_, err = f.wallet.SendTransaction(newBlock, randomAddr, big.NewInt(42069))
				if err != nil {
					log.Error().Err(err).Str("Fan", f.Name).Str("ID", f.ID).Msg("Error sending transaction")
				}
			}
		}
	}
}

// ReceiveBlock receives a new block from the chain, and acts according to the fan's intensity level
func (f *Fan) ReceiveBlock(block *types.Block) {
	if !f.IsSearching() {
		return
	}
	f.blockChan <- block
	log.Trace().Str("Fan ID", f.ID).Str("Crazed Level", f.CrazedLevel.Name).Str("Fan Name", f.Name).Msg("Received new header")
}

// Stop stops the fan's search
func (f *Fan) Stop() {
	f.stopButton <- struct{}{}
}

// IsSearching returns whether or not the fan is currently searching
func (f *Fan) IsSearching() bool {
	f.searchingMu.Lock()
	defer f.searchingMu.Unlock()
	return f.currentlySearching
}
