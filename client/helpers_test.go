package client_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/kalverra/crazed-nft-fans/client"
	"github.com/stretchr/testify/require"
)

func TestPrivToAddr(t *testing.T) {
	t.Parallel()

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err, "Error generating new private key")
	addr, err := client.PrivateKeyToAddress(privateKey)
	require.NoError(t, err, "Error converting key to address")
	require.NotNil(t, addr, "Nil address")
	require.NotEmpty(t, addr, "Empty address")
}

func TestConversions(t *testing.T) {
	t.Parallel()

	wei := big.NewInt(1500000000000000000)
	gwei := big.NewFloat(1500000000)
	eth := big.NewFloat(1.5)

	weiVal := client.EtherToWei(eth)
	require.Equal(t, 0, weiVal.Cmp(wei), "1.5 Ether converted incorrectly")

	ethVal := client.WeiToEther(wei)
	require.Equal(t, 0, ethVal.Cmp(eth), "1.5 Ether converted incorrectly")

	gweiVal := client.WeiToGwei(wei)
	require.Equal(t, 0, gweiVal.Cmp(gwei), "1.5 Ether converted incorrectly")
}
