// Package config defines the configuration of the fans and the blockchain they're pointing to
package config

import (
	"crypto/ecdsa"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/convert"
)

// Current holds the current project's config
var Current *Config

// Config details the config for the project
type Config struct {
	HTTP    string `envconfig:"http_url" default:"http://localhost:8545"` // HTTP URL of the chain
	WS      string `envconfig:"ws_url" default:"ws://localhost:8546"`     // Websocket URL of the chain
	ChainID uint64 `envconfig:"chain_id" default:"1337"`                  // ID of the chain
	// Funding Key is the main key to fund fans from. Default is the default used by geth, hardhat, ganache, etc...
	FundingKey        string  `envconfig:"funding_key" default:"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"`
	PeakGasPriceGwei  float64 `envconfig:"peak_gas_price" default:"100"` // Target gas price in Gwei
	FloorGasPriceGwei float64 `envconfig:"floor_gas_price" default:"10"` // Target gas price in Gwei
	LogLevel          string  `envconfig:"log_level" default:"debug"`

	FundingPrivateKey *ecdsa.PrivateKey `ignored:"true"` // Transformed private key
	FundingAddress    common.Address    `ignored:"true"` // Transformed private key to address
	BigChainID        *big.Int          `ignored:"true"` // ChainID in big.Int format
	PeakGasPriceWei   *big.Int          `ignored:"true"` // Target gas price in Wei
	FloorGasPriceWei  *big.Int          `ignored:"true"` // Floor gas price in Wei
}

// ReadConfig reads in the project config in from env vars
func ReadConfig() error {
	var err error
	var conf Config
	if err = envconfig.Process("", &conf); err != nil {
		return err
	}
	if err = InitLogging(conf.LogLevel); err != nil {
		return err
	}

	conf.FundingPrivateKey, err = crypto.HexToECDSA(conf.FundingKey)
	if err != nil {
		return err
	}

	conf.FundingAddress = crypto.PubkeyToAddress(conf.FundingPrivateKey.PublicKey)
	conf.PeakGasPriceWei = convert.GweiToWei(big.NewFloat(conf.PeakGasPriceGwei))
	conf.FloorGasPriceWei = convert.GweiToWei(big.NewFloat(conf.FloorGasPriceGwei))
	conf.BigChainID = new(big.Int).SetUint64(conf.ChainID)
	Current = &conf
	return err
}

// InitLogging initializes logging based on the passed in level
func InitLogging(logLevel string) error {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(level)
	return nil
}
