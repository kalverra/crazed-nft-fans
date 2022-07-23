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

var toAddress string = "0x024c0763F8b55972Cd4a0349c79833bA9e3B2279"

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	os.Exit(m.Run())
}

func TestBadClient(t *testing.T) {
	err := os.Setenv("WS_URL", "ws://fake.url")
	require.NoError(t, err, "Error setting fake url for test")

	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	_, err = client.NewClient()
	require.Error(t, err, "Expected a fake URL to make the eth client throw an error")

	err = os.Unsetenv("WS_URL")
	require.NoError(t, err, "Error un-setting fake url")
}

func TestConnectClient(t *testing.T) {
	err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	client, err := client.NewClient()
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, client, "Nil client")
}

func TestSendTx(t *testing.T) {
	to := common.HexToAddress(toAddress)
	ethClient, err := client.NewClient()
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, ethClient, "Nil client")
	fundingAddr, err := client.PrivateKeyToAddress(config.Current.FundingPrivateKey)
	require.NoError(t, err, "Error getting funding key address")

	startingFromBalance, err := ethClient.BalanceAt(fundingAddr)
	require.NoError(t, err, "Error retrieving balance")
	require.Equal(t, startingFromBalance.Cmp(big.NewInt(0)), 1, "Funding balance is 0 or less")
	startingToBalance, err := ethClient.BalanceAt(to)
	require.NoError(t, err, "Error retrieving balance")

	hash, err := ethClient.SendTransaction(config.Current.FundingPrivateKey, to, big.NewInt(0), big.NewFloat(100))
	require.NoError(t, err, "Error sending transaction")

	ctxt, cancel := context.WithTimeout(context.Background(), time.Second*15)
	confirmed, err := ethClient.ConfirmTransaction(ctxt, hash)
	require.NoError(t, err, "Error confirming tx")
	cancel()
	require.True(t, confirmed, "Transaction not confirmed")

	finalFromBalance, err := ethClient.BalanceAt(fundingAddr)
	require.NoError(t, err, "Error retrieving balance")
	finalToBalance, err := ethClient.BalanceAt(to)
	require.NoError(t, err, "Error retrieving balance")

	require.Equal(t, 1, startingFromBalance.Cmp(finalFromBalance),
		"Starting From balance '%s' should be less than the final balance '%s'",
		client.WeiToEther(startingFromBalance).String(), client.WeiToEther(finalFromBalance).String())
	require.Equal(t, -1, startingToBalance.Cmp(finalToBalance),
		"Starting To balance '%s' should be less than the final balance '%s'",
		client.WeiToEther(startingToBalance).String(), client.WeiToEther(finalToBalance).String())
}
