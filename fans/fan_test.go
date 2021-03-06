//go:build integration
// +build integration

package fans_test

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
)

func TestMain(m *testing.M) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if err := config.ReadConfig(); err != nil {
		log.Fatal().Err(err).Msg("Error reading config")
	}
	os.Exit(m.Run())
}

func TestNewFan(t *testing.T) {
	t.Parallel()

	fan, err := fans.NewFan()
	require.NoError(t, err, "Error creating new fan")

	require.NotEmpty(t, fan.ID, "Empty fan ID")
	require.NotEmpty(t, fan.Name, "Empty fan Name")
	require.NotEmpty(t, fan.PrivateKey, "Empty fan PrivateKey")
	require.NotEmpty(t, fan.Address, "Empty fan Address")
}

func TestStopSearch(t *testing.T) {
	t.Parallel()

	fan, err := fans.NewFan()
	require.NoError(t, err, "Error creating new fan")

	go fan.Search()

	time.Sleep(time.Millisecond)
	require.True(t, fan.IsSearching(), "Fan should have an active searching status")
	fan.StopSearch()
	time.Sleep(time.Millisecond)
	require.False(t, fan.IsSearching(), "Fan should no longer be searching")
}
