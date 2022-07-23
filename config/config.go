// Package config defines the config for the project
package config

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
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
	CrazedLevel       int               `envconfig:"crazed_level" default:"0"` // Crazed level for the fans
	FundingPrivateKey *ecdsa.PrivateKey `ignored:"true"`                       // Transformed private key
	BigChainID        *big.Int          `ignored:"true"`                       // ChainID in big.Int format
}

// ReadConfig reads in the project config in from env vars
func ReadConfig() error {
	var err error
	var conf Config
	if err = envconfig.Process("", &conf); err != nil {
		return err
	}
	if conf.CrazedLevel > 5 || conf.CrazedLevel < 0 {
		log.Warn().Int("Selected", conf.CrazedLevel).Msg("Invalid Crazed Level selected. Defaulting to Mixed")
		conf.CrazedLevel = 0
	}
	conf.BigChainID = new(big.Int).SetUint64(conf.ChainID)
	conf.FundingPrivateKey, err = crypto.HexToECDSA(conf.FundingKey)
	if err != nil {
		return err
	}
	Current = &conf
	return err
}

// GetCrazedLevel retrieves the current crazed level, processing Mixed if necessary
func (c *Config) GetCrazedLevel() int {
	crazedLevel := c.CrazedLevel
	if crazedLevel == 0 {
		crazedLevel = rand.Intn(5) + 1
	}
	return crazedLevel
}

// CrazedLevelMappings maps the crazed level to the adjectives
var CrazedLevelMappings = map[int]string{
	0: "Mixed",
	1: "Indifferent",
	2: "Curious",
	3: "Interested",
	4: "Obsessed",
	5: "Manic",
}
