//go:build integration
// +build integration

package client_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

var (
	toAddress string = "0x024c0763F8b55972Cd4a0349c79833bA9e3B2279"
	conf      *config.Config
)

func init() {
	var err error
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	conf, err = config.ReadConfig()
	if err != nil {
		log.Fatal().Msg("Error reading config")
	}
}

func TestConnectClient(t *testing.T) {
	t.Parallel()

	_, err := client.NewClient("ws://fakeurl.io")
	require.Error(t, err, "Expected a fake URL to make the eth client throw an error")

	conf, err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	client, err := client.NewClient(conf.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, client, "Nil client")
}

func TestSendTx(t *testing.T) {
	conf, err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")

	ethClient, err := client.NewClient(conf.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, ethClient, "Nil client")
	fundingAddr, err := client.PrivateKeyToAddress(conf.FundingPrivateKey)
	require.NoError(t, err, "Error getting funding key address")
	fundingBalance, err := ethClient.BalanceAt(fundingAddr)
	require.NoError(t, err, "Error retrieving funding key balance")
	require.Equal(t, fundingBalance.Cmp(big.NewInt(0)), 1, "Funding balance is 0 or less")

	hash, err := ethClient.SendTransaction(conf.FundingPrivateKey, common.HexToAddress(toAddress), 1, big.NewInt(0), big.NewFloat(1))
	require.NoError(t, err, "Error sending transaction")

	ctxt, cancel := context.WithTimeout(context.Background(), time.Second*5)
	confirmed, err := ethClient.ConfirmTransaction(ctxt, hash)
	require.NoError(t, err, "Error confirming tx")
	cancel()
	require.True(t, confirmed, "Transaction not confirmed")
}
