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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

var toAddress = common.HexToAddress("0x024c0763F8b55972Cd4a0349c79833bA9e3B2279")

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading initial config file")
	}
	err = client.NewTransactionTracker()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing transaction tracker")
	}
	os.Exit(m.Run())
}

func TestBadClient(t *testing.T) {
	t.Parallel()

	_, err := client.NewClient("ws://fake.url")
	require.Error(t, err, "Expected a fake URL to make the eth client throw an error")
}

func TestConnectClient(t *testing.T) {
	t.Parallel()

	client, err := client.NewClient(config.Current.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, client, "Nil client")
}

func TestSendTx(t *testing.T) {
	t.Parallel()

	ethClient, err := client.NewClient(config.Current.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, ethClient, "Nil client")
	fundingAddr, err := client.PrivateKeyToAddress(config.Current.FundingPrivateKey)
	require.NoError(t, err, "Error getting funding key address")

	startingFromBalance, err := ethClient.BalanceAt(fundingAddr)
	require.NoError(t, err, "Error retrieving balance")
	require.Equal(t, startingFromBalance.Cmp(big.NewInt(0)), 1, "Funding balance is 0 or less")
	startingToBalance, err := ethClient.BalanceAt(toAddress)
	require.NoError(t, err, "Error retrieving balance")

	fundingNonce := client.GlobalTransactionTracker.FirstAvailableNonce(fundingAddr)
	hash, err := ethClient.SendTransaction(config.Current.FundingPrivateKey, toAddress, fundingNonce, big.NewInt(0), big.NewFloat(100))
	require.NoError(t, err, "Error sending transaction")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	confirmed, err := ethClient.ConfirmTransaction(ctx, fundingAddr, hash)
	require.NoError(t, err, "Error confirming tx")
	cancel()
	require.True(t, confirmed, "Transaction not confirmed")

	finalFromBalance, err := ethClient.BalanceAt(fundingAddr)
	require.NoError(t, err, "Error retrieving balance")
	finalToBalance, err := ethClient.BalanceAt(toAddress)
	require.NoError(t, err, "Error retrieving balance")

	require.Equal(t, 1, startingFromBalance.Cmp(finalFromBalance),
		"Starting From balance '%s' should be less than the final balance '%s'",
		client.WeiToEther(startingFromBalance).String(), client.WeiToEther(finalFromBalance).String())
	require.Equal(t, -1, startingToBalance.Cmp(finalToBalance),
		"Starting To balance '%s' should be less than the final balance '%s'",
		client.WeiToEther(startingToBalance).String(), client.WeiToEther(finalToBalance).String())
}

func TestSubscription(t *testing.T) {
	t.Parallel()

	client, err := client.NewClient(config.Current.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, client, "Nil client")
	newBlocks := make(chan *types.Header)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	sub, err := client.SubscribeNewBlocks(ctx, newBlocks)
	require.NoError(t, err, "Error subscribing to new blocks")

testLoop:
	for {
		select {
		case err = <-sub.Err():
			break testLoop
		case <-newBlocks:
			break testLoop
		case <-ctx.Done():
			break testLoop
		}
	}

	require.NoError(t, err, "Error while subscribing to new blocks")
}

func TestConfirm(t *testing.T) {
	t.Parallel()

	ethClient, err := client.NewClient(config.Current.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, ethClient, "Nil client")

	fundingAddr, err := client.PrivateKeyToAddress(config.Current.FundingPrivateKey)
	require.NoError(t, err, "Error getting funding key address")
	fundingNonce := client.GlobalTransactionTracker.FirstAvailableNonce(fundingAddr)
	hash, err := ethClient.SendTransaction(config.Current.FundingPrivateKey, toAddress, fundingNonce, big.NewInt(0), big.NewFloat(100))
	require.NoError(t, err, "Error sending transaction")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	confirmed, err := ethClient.ConfirmTransaction(ctx, fundingAddr, hash)
	require.NoError(t, err, "Error confirming tx")
	cancel()
	require.True(t, confirmed, "Transaction not confirmed")
}
