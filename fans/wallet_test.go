package fans_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
)

func TestNewWallet(t *testing.T) {
	t.Parallel()

	client, err := ethclient.Dial(config.Current.WS)
	require.NoError(t, err, "Error dialing client")

	wallet, err := fans.NewWallet("test", "test", client, config.Current.GetCrazedLevel())
	require.NoError(t, err, "Error creating new wallet")
	require.NotNil(t, wallet, "Wallet should not be nil")
}
