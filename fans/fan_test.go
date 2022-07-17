package fans_test

import (
	"testing"

	"github.com/kalverra/crazed-nft-fans/fans"
	"github.com/stretchr/testify/require"
)

func TestNewFan(t *testing.T) {
	fan, err := fans.NewFan()

	require.NoError(t, err, "Error creating new fan")
	require.NotEmpty(t, fan.ID, "Empty fan ID")
	require.NotEmpty(t, fan.Name, "Empty fan Name")
	require.NotEmpty(t, fan.PrivateKey, "Empty fan PrivateKey")
	require.NotEmpty(t, fan.Address, "Empty fan Address")
}
