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
	t.Setenv("HTTP_URL", http)
	t.Setenv("WS_URL", ws)
	t.Setenv("CHAIN_ID", fmt.Sprint(chainId))

	err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")

	require.Equal(t, http, config.Current.HTTP)
	require.Equal(t, ws, config.Current.WS)
	require.Equal(t, chainId, config.Current.ChainID)
}

func TestBadRead(t *testing.T) {
	t.Setenv("CHAIN_ID", "badValue")
	err := config.ReadConfig()
	require.Error(t, err, "Config should have shown an error setting CHAIN_ID='badValue'")
}

func TestCrazedLevel(t *testing.T) {
	t.Setenv("CRAZED_LEVEL", "0")
	err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, 0, config.Current.CrazedLevel)
	require.NotEqual(t, 0, config.Current.GetCrazedLevel())

	t.Setenv("CRAZED_LEVEL", "7")
	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, 0, config.Current.CrazedLevel, "Invalid crazed level should have been changed to 0")
}

func TestBadKey(t *testing.T) {
	t.Setenv("FUNDING_KEY", "badKey")
	err := config.ReadConfig()
	require.Error(t, err, "Bad funding key should have thrown an error")
}
