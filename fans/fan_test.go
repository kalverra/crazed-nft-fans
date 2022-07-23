package fans_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/fans"
)

func TestNewFan(t *testing.T) {
	t.Parallel()

	fan, err := fans.NewFan()

	require.NoError(t, err, "Error creating new fan")
	require.NotEmpty(t, fan.ID, "Empty fan ID")
	require.NotEmpty(t, fan.Name, "Empty fan Name")
	require.NotEmpty(t, fan.PrivateKey, "Empty fan PrivateKey")
	require.NotEmpty(t, fan.Address, "Empty fan Address")
}
