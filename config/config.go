// Package config defines the config for the project
package config

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kelseyhightower/envconfig"
)

// Current holds the current project's config
var Current *Config

// Config details the config for the project
type Config struct {
	HTTP    string `envconfig:"http_url" default:"http://localhost:8545"` // HTTP URL of the chain
	WS      string `envconfig:"ws_url" default:"ws://localhost:8546"`     // Websocket URL of the chain
	ChainID uint64 `envconfig:"chain_id" default:"1337"`                  // ID of the chain
	// Funding Key is the main key to fund fans from. Default is the default used by geth, hardhat, ganache, etc...
	FundingKey        string            `envconfig:"funding_key" default:"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"`
	FundingPrivateKey *ecdsa.PrivateKey `ignored:"true"` // Transformed private key
	BigChainID        *big.Int          `ignored:"true"` // ChainID in big.Int format
}

// ReadConfig reads in the project config in from env vars
func ReadConfig() (*Config, error) {
	var err error
	var conf Config
	if err = envconfig.Process("", &conf); err != nil {
		return nil, err
	}
	conf.BigChainID = new(big.Int).SetUint64(conf.ChainID)
	conf.FundingPrivateKey, err = crypto.HexToECDSA(conf.FundingKey)
	return &conf, err
}
