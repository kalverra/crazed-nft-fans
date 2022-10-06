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
	"github.com/ethereum/go-ethereum/crypto"
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
	os.Exit(m.Run())
}

func TestBadClient(t *testing.T) {
	t.Parallel()

	_, err := client.NewWallet(config.Current.FundingPrivateKey, "ws://fake.url")
	require.Error(t, err, "Expected a fake URL to make the eth client throw an error")
}

func TestConnectClient(t *testing.T) {
	t.Parallel()

	wallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, wallet, "Nil client")
}

func TestSameWallet(t *testing.T) {
	t.Parallel()

	firstWallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, firstWallet, "Nil wallet")
	secondWallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, secondWallet, "Nil wallet")
	require.Equal(t, firstWallet, secondWallet, "Wallets should be the same")
}

func TestDifferentWallet(t *testing.T) {
	t.Parallel()

	firstKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating key")
	secondKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating key")

	firstWallet, err := client.NewWallet(firstKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, firstWallet, "Nil wallet")
	secondWallet, err := client.NewWallet(secondKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, secondWallet, "Nil wallet")
	require.NotEqual(t, firstWallet, secondWallet, "Wallets should not be the same")
}

func TestSubscription(t *testing.T) {
	t.Parallel()

	wallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, wallet, "Nil wallet")
	newBlocks := make(chan *types.Header)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	sub, err := wallet.SubscribeNewBlocks(ctx, newBlocks)
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

func TestTransaction(t *testing.T) {
	t.Parallel()

	fundingWallet, err := client.NewWallet(config.Current.FundingPrivateKey, config.Current.WS)
	require.NoError(t, err, "Error creating wallet")
	require.NotNil(t, fundingWallet, "Nil wallet")
	toKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating key")
	toAddress, err := client.PrivateKeyToAddress(toKey)

	fundingBalanceBefore, err := fundingWallet.Balance()
	require.NoError(t, err, "Error getting balance")
	toBalanceBefore, err := fundingWallet.BalanceAt(toAddress)
	require.NoError(t, err, "Error getting balance")

	require.Equal(t, fundingBalanceBefore.Cmp(big.NewInt(0)), 1, "FromAddress balance should be more than 0")
	require.Equal(t, toBalanceBefore.Cmp(big.NewInt(0)), 0, "ToAddress balance should start at 0")

	fundingWallet.SendTransaction(time.Second*5, big.NewFloat(1.125), toAddress, big.NewInt(1), big.NewFloat(.01))

	fundingBalanceAfter, err := fundingWallet.Balance()
	require.NoError(t, err, "Error getting balance")
	toBalanceAfter, err := fundingWallet.BalanceAt(toAddress)
	require.NoError(t, err, "Error getting balance")

	require.Equal(t, fundingBalanceBefore.Cmp(fundingBalanceAfter), 1, "Funding balance should be less after transaction")
	require.Equal(t, toBalanceBefore.Cmp(toBalanceAfter), -1, "To balance should be more after transaction")
}
