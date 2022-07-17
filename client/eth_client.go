package client

import "github.com/ethereum/go-ethereum/ethclient"

// EthClient wraps the standard Ethereum client
type EthClient struct {
	InnerClient *ethclient.Client
}

// NewClient produces a new client to connect to the blockchain
func NewClient(url string) (*EthClient, error) {
	ethClient, err := ethclient.Dial(url)
	return &EthClient{
		InnerClient: ethClient,
	}, err
}
