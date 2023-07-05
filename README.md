<div align="center">

# Crazed NFT Fans

[![Go Reference](https://pkg.go.dev/badge/github.com/kalverra/crazed-nft-fans.svg)](https://pkg.go.dev/github.com/kalverra/crazed-nft-fans)
[![Tests](https://github.com/kalverra/crazed-nft-fans/actions/workflows/integration-test.yaml/badge.svg)](https://github.com/kalverra/crazed-nft-fans/actions/workflows/integration-test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kalverra/crazed-nft-fans)](https://goreportcard.com/report/github.com/kalverra/crazed-nft-fans)
[![codecov](https://codecov.io/gh/kalverra/crazed-nft-fans/branch/main/graph/badge.svg)](https://codecov.io/gh/kalverra/crazed-nft-fans)
[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/kalverra/crazed-nft-fans/main/LICENSE)

</div>

A new NFT has dropped, but its location is a secret! Fans have become crazed, and are sending funds to random wallets and calling gas guzzling contracts as fast as they can to hopefully snag it.

Using the fans on a simulated network (like [geth in dev mode](https://geth.ethereum.org/docs/getting-started/dev-mode)) can help you emulate network congestion events.

If using a simulated geth instance, you might find it helpful to [turn on metrics](https://geth.ethereum.org/docs/interface/metrics).

## Configure

Environment variables are used to configure everything about the crazed fans. You can set them in a `.env` file, or export them in your shell. See the `example.env` file for an example.

```sh
HTTP_URL="http://localhost:8545" # HTTP URL of the chain to run on
WS_URL="ws://localhost:8546" # WS URL of the chain to run on
CHAIN_ID="1337" # ID of the chain to run on
FUNDING_KEY="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80" # Private key of the funding address
TARGET_GAS_PRICE="1000000000" # Gas price to target (in Gwei) as the peak on chain price.
```

## Run


## Emulating a Network Congestion Event

This is the tricky bit, you can't 

## Test

`make test`

to run basic tests in standard go format, or

`make test_fancy`

to run basic tests with a prettier output, or

`make test_integration`

to launch a simulated geth node to run all possible tests.

### Can I use this to cause chaos on ethereum mainnet or testnets?

Maybe? But I wouldn't recommend it, unless you are looking for a way to become very poor, very fast. There's far more fun ways to do that anyway.
