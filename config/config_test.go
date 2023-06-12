package config_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
)

func TestMain(m *testing.M) {
	err := config.InitLogging("debug")
	if err != nil {
		log.Fatal(err)
	}
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
	t.Setenv("CRAZED_LEVEL", "Obsessed")
	err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, "Obsessed", config.Current.CrazedLevel)
	require.Equal(t, "Obsessed", config.Current.GetCrazedLevel().Name)

	t.Setenv("CRAZED_LEVEL", "Invalid")
	err = config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, "Mixed", config.Current.CrazedLevel, "Invalid crazed level should have been changed to Mixed")
}

func TestMixedCrazedLevel(t *testing.T) {
	t.Setenv("CRAZED_LEVEL", "Mixed")
	err := config.ReadConfig()
	require.NoError(t, err, "Error reading config")
	require.Equal(t, "Mixed", config.Current.CrazedLevel)
	require.NotEqual(t, "Mixed", config.Current.GetCrazedLevel().Name)
}

func TestBadKey(t *testing.T) {
	t.Setenv("FUNDING_KEY", "badKey")
	err := config.ReadConfig()
	require.Error(t, err, "Bad log level should have thrown an error")
}

func TestBadLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "badLog")
	err := config.ReadConfig()
	require.Error(t, err, "Bad funding key should have thrown an error")
}
