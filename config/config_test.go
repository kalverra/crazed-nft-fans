package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
)

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	os.Exit(m.Run())
}

func TestRead(t *testing.T) {
	http := "http://url.com"
	ws := "ws://url.com/ws"
	chainId := uint64(420)
	err := os.Setenv("HTTP_URL", http)
	require.NoError(t, err, "Error setting env var")
	os.Setenv("WS_URL", ws)
	require.NoError(t, err, "Error setting env var")
	os.Setenv("CHAIN_ID", fmt.Sprint(chainId))
	require.NoError(t, err, "Error setting env var")

	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")

	require.Equal(t, http, config.Current.HTTP)
	require.Equal(t, ws, config.Current.WS)
	require.Equal(t, chainId, config.Current.ChainID)
}

func TestBadRead(t *testing.T) {
	err := os.Setenv("CHAIN_ID", "badValue")
	require.NoError(t, err, "Error setting env var")
	err = config.ReadConfig()
	require.Error(t, err, "Config should have shown an error setting CHAIN_ID='badValue'")
	err = os.Unsetenv("CHAIN_ID")
	require.NoError(t, err, "Error un-setting env var")
}

func TestCrazedLevel(t *testing.T) {
	err := os.Setenv("CRAZED_LEVEL", "0")
	require.NoError(t, err, "Error setting env var")
	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, 0, config.Current.CrazedLevel)
	require.NotEqual(t, 0, config.Current.GetCrazedLevel())

	err = os.Setenv("CRAZED_LEVEL", "7")
	require.NoError(t, err, "Error setting env var")
	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, 0, config.Current.CrazedLevel, "Invalid crazed level should have been changed to 0")
}
