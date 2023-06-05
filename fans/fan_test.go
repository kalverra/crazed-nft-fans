package fans_test

import (
	"testing"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"

	"github.com/stretchr/testify/require"
)

func TestNewFan(t *testing.T) {
	fan, err := fans.New(config.Manic)
	require.NoError(t, err, "Error creating new fan")
	require.NotNil(t, fan, "Fan should not be nil")
}
