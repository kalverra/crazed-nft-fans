package client_test

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/kalverra/crazed-nft-fans/config"
)

func TestDoubleInitialization(t *testing.T) {
	t.Parallel()

	err := client.NewTransactionTracker()
	require.Error(t, err, "Trying to initialize tracker more than once should throw an error.")
}

func TestCompleteNonTrackedTx(t *testing.T) {
	t.Parallel()

	err := client.GlobalTransactionTracker.CompletedTransaction(common.HexToAddress("0x0"), common.HexToHash("0x0"))
	require.Error(t, err, "Completing a non-tracked address should result in an error")

	fundingHelper(t, config.Current.FundingPrivateKey, common.HexToAddress("0x0"), big.NewFloat(1))

	fundingAddr, err := client.PrivateKeyToAddress(config.Current.FundingPrivateKey)
	require.NoError(t, err, "Error getting funding key address")
	err = client.GlobalTransactionTracker.CompletedTransaction(fundingAddr, common.HexToHash("0x0"))
	require.Error(t, err, "Completing a nonexistent tx should give an error")
}

func TestNewTx(t *testing.T) {
	t.Parallel()

	newPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating key")
	address, err := client.PrivateKeyToAddress(newPrivateKey)
	require.NoError(t, err, "Error converting to address")

	fundingHelper(t, config.Current.FundingPrivateKey, address, big.NewFloat(5))

	toKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating key")
	toAddress, err := client.PrivateKeyToAddress(toKey)
	require.NoError(t, err, "Error converting to address")

	fundingHelper(t, newPrivateKey, toAddress, big.NewFloat(1))
}

func fundingHelper(t *testing.T, fromPrivateKey *ecdsa.PrivateKey, toAddress common.Address, amount *big.Float) {
	t.Helper()

	ethClient, err := client.NewClient(config.Current.WS)
	require.NoError(t, err, "Error connecting client")
	require.NotNil(t, ethClient, "Nil client")
	fromAddr, err := client.PrivateKeyToAddress(fromPrivateKey)
	require.NoError(t, err, "Error getting funding key address")

	fundingNonce := client.GlobalTransactionTracker.FirstAvailableNonce(fromAddr)
	txHash, err := ethClient.SendTransaction(
		fromPrivateKey,
		toAddress,
		fundingNonce,
		big.NewInt(0),
		amount,
	)
	require.NoError(t, err, "Error while sending tx")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	confirmed, err := ethClient.ConfirmTransaction(ctx, fromAddr, txHash)
	require.NoError(t, err, "Error confirming tx")
	require.True(t, confirmed, "Tx not confirmed")
	cancel()
}
