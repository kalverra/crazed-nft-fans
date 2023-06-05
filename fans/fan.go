package fans

import (
	"crypto/ecdsa"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/rs/zerolog/log"
)

// Fan is an NFT fan that will search for NFTs
type Fan struct {
	ID          uint64
	Name        string
	Address     string
	PrivateKey  *ecdsa.PrivateKey
	CrazedLevel *config.CrazedLevel

	stopButton          chan struct{}
	currentlySearching  bool
	pendingTransactions []string
	headerMu            sync.Mutex
}

// New creates a new fan at a supplied crazed level
func New(level *config.CrazedLevel) (*Fan, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &Fan{
		ID:                  rand.Uint64(),
		Name:                generateName(),
		Address:             crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
		PrivateKey:          privateKey,
		CrazedLevel:         level,
		stopButton:          make(chan struct{}),
		currentlySearching:  false,
		pendingTransactions: []string{},
	}, nil
}

// Search begins the fan's search
func (f *Fan) Search() {
	f.currentlySearching = true
	log.Info().Uint64("ID", f.ID).Str("Name", f.Name).Msg("Fan Searching!")
	go func() {
		<-f.stopButton
		log.Info().Uint64("ID", f.ID).Str("Name", f.Name).Msg("Stopping fan")
		f.currentlySearching = false
	}()
}

// ReceiveHeader receives a new header from the chain, and acts according to the fan's intensity level
func (f *Fan) ReceiveHeader(header *types.Header) {
	f.headerMu.Lock()
	defer f.headerMu.Unlock()
	if !f.IsSearching() {
		return
	}
	log.Trace().Uint64("Fan ID", f.ID).Str("Crazed Level", f.CrazedLevel.Name).Str("Fan Name", f.Name).Msg("Received new header")
}

// Stop stops the fan's search
func (f *Fan) Stop() {
	f.stopButton <- struct{}{}
}

func (f *Fan) IsSearching() bool {
	return f.currentlySearching
}
