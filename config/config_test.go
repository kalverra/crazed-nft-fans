package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	t.Parallel()

	http := "http://url.com"
	ws := "ws://url.com/ws"
	chainId := uint64(420)
	os.Setenv("HTTP_URL", http)
	os.Setenv("WS_URL", ws)
	os.Setenv("CHAIN_ID", fmt.Sprint(chainId))

	conf, err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")

	require.Equal(t, http, conf.HTTP)
	require.Equal(t, ws, conf.WS)
	require.Equal(t, chainId, conf.ChainID)
}
