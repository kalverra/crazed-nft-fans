//go:build integration
// +build integration

package client_test

import (
	"os"
	"testing"

	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
}

func TestConnectClient(t *testing.T) {
	_, err := client.NewClient("ws://fakeurl.io")
	require.Error(t, err, "Expected a fake URL to make the eth client throw an error")
}
