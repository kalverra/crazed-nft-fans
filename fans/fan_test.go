//go:build integration
// +build integration

package fans_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/convert"
	"github.com/kalverra/crazed-nft-fans/fans"
)

func TestMain(m *testing.M) {
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading config")
	}
	m.Run()
}

func TestTransaction(t *testing.T) {
	client, err := ethclient.Dial(config.Current.WS)
	require.NoError(t, err, "Error dialing client")
	fan, err := fans.New(client)
	require.NoError(t, err, "Error creating fan")
	pendingNonce, err := client.PendingNonceAt(context.Background(), config.Current.FundingAddress)
	require.NoError(t, err, "Error getting pending nonce")
	err = fan.Fund(convert.EtherToWei(big.NewFloat(0.1)), pendingNonce, time.Second*10)
	require.NoError(t, err, "Error funding fan")

	header, err := client.HeaderByNumber(context.Background(), nil)
	hash, err := fan.SendRandomTransaction(header.BaseFee)
	require.NoError(t, err, "Error sending random transaction")

	// Wait for transaction to be mined
	check, timeout := time.NewTicker(time.Millisecond*500), time.After(time.Second*10)
	defer check.Stop()
	for {
		select {
		case <-check.C:
			receipt, err := client.TransactionReceipt(context.Background(), hash)
			require.NoError(t, err, "Error getting transaction receipt")
			if receipt.Status == 1 {
				return
			}
		case <-timeout:
			require.Fail(t, "Transaction timed out")
		}
	}
}
