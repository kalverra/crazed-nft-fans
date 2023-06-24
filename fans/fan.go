package fans

import (
	"context"
	"crypto/ecdsa"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
)

// Fan is an NFT fan that will search for NFTs
type Fan struct {
	ID          string
	Name        string
	Address     string
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
func New(president *President, level *config.CrazedLevel) (*Fan, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	fan := &Fan{
		ID:          xid.New().String(),
		Name:        generateName(),
		Address:     crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
		PrivateKey:  privateKey,
		CrazedLevel: level,

		stopButton:         make(chan struct{}),
		currentlySearching: false,
		blockChan:          make(chan *types.Block),
	}
	fan.wallet, err = NewFanWallet(fan)
	if err != nil {
		return nil, err
	}
	if president != nil { // nil president for testing
		fan.client = president.Client
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
			f.wallet.UpdatePendingTxs(newBlock)
			if len(f.wallet.PendingTransactions()) < f.CrazedLevel.MaxPendingTransactions {
				initialTipCap, err := f.client.SuggestGasTipCap(context.Background())
				if err != nil {
					log.Error().Err(err).Msg("Error getting suggested gas tip cap")
				}
				f.wallet.SendTransaction(f.pendingNonce, initialTipCap)
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
