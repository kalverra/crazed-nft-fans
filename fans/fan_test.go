//go:build integration
// +build integration

package fans_test

import (
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
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

func TestFunding(t *testing.T) {
	t.Parallel()

	fans := president.Fans()
	president.FundFans(big.NewFloat(1))
	for _, fan := range fans {
		funds, err := fan.Wallet.Balance()
		require.NoError(t, err, "Error getting wallet balance")
		ethBalance := client.WeiToEther(funds)
		require.GreaterOrEqual(t, ethBalance.Cmp(big.NewFloat(1)), 0, "Not enough funds in wallet, only %s in wallet", ethBalance.String())
	}
}

func TestNewFan(t *testing.T) {
	t.Parallel()

	fan, err := fans.NewFan()
	require.NoError(t, err, "Error creating new fan")

	require.NotEmpty(t, fan.ID, "Empty fan ID")
	require.NotEmpty(t, fan.Name, "Empty fan Name")
	require.NotEmpty(t, fan.PrivateKey, "Empty fan PrivateKey")
	require.NotEmpty(t, fan.Address, "Empty fan Address")
}

func TestStopSearch(t *testing.T) {
	t.Parallel()

	err := president.NewFans(1)
	require.NoError(t, err, "Error creating new fans")
	president.FundFans(big.NewFloat(1))

	searchingFan := president.Fans()[0]
	err = searchingFan.Search()
	require.NoError(t, err, "Error searching")

	require.True(t, searchingFan.IsSearching(), "Fan should have an active searching status")
	searchingFan.StopSearch()

	time.Sleep(time.Millisecond) // Yuck, but necessary/intended functionality
	require.False(t, searchingFan.IsSearching(), "Fan should no longer be searching")
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
