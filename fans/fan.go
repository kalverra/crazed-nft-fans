// Package fans details the actions of each crazed fan
package fans

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	"github.com/kalverra/crazed-nft-fans/client"
)

// Fan is an NFT fan that searches for the NFT by incessantly bumping gas
type Fan struct {
	ID          string
	Name        string
	Client      *client.EthClient
	PrivateKey  *ecdsa.PrivateKey
	Address     common.Address
	CrazedLevel int
}

// NewFan generates a new fan
func NewFan() (*Fan, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	address, err := client.PrivateKeyToAddress(privateKey)
	if err != nil {
		return nil, err
	}

	return &Fan{
		ID:         uuid.New().String(),
		Name:       newName(),
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}
