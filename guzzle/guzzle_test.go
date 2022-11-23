//go:build integration
// +build integration

package guzzle_test

import (
	"math/big"
	"os"
	"testing"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/test-go/testify/require"
)

var president *fans.FanPresident

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading initial config file")
	}
	president, err = fans.NewPresident()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating new fan president")
	}
	err = president.NewFans(5)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating new fans")
	}
	os.Exit(m.Run())
}

func TestGasGuzzle(t *testing.T) {
	t.Parallel()

	err := president.DeployGasGuzzlers(1)
	require.NoError(t, err, "Error deploying gas guzzlers")
	guzzlers := president.GasGuzzlers()
	require.NotEmpty(t, guzzlers, "No guzzlers found")
	president.FundFans(big.NewFloat(1))

	fan := president.Fans()[0]
	beforeBal, err := fan.Wallet.Balance()
	require.NoError(t, err, "Error getting fan balance")
	require.Equal(t, 1, beforeBal.Cmp(big.NewInt(0)), "Fan balance should be more than 0")
	err = fan.Guzzle(guzzlers[0], 50_000)
	require.NoError(t, err, "Error guzzling gas")
	afterBal, err := fan.Wallet.Balance()
	require.NoError(t, err, "Error getting fan balance")
	require.Equal(t, 1, beforeBal.Cmp(afterBal), "Fan balance should be less after burning gas")

	err = fan.Guzzle(guzzlers[0], 50_000_000)
	require.Error(t, err, "Guzzle should throw an error for guzzling more gas than block limit")
}
