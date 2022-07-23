package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
)

func TestRead(t *testing.T) {
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
