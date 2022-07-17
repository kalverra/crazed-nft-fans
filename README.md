# Crazed NFT Fans

[![Go Reference](https://pkg.go.dev/badge/github.com/kalverra/crazed-nft-fans.svg)](https://pkg.go.dev/github.com/kalverra/crazed-nft-fans)
[![Tests](https://github.com/kalverra/crazed-nft-fans/workflows/test/badge.svg)](https://github.com/kalverra/crazed-nft-fans/actions?workflow=test)
[![Go Report Card](https://goreportcard.com/badge/github.com/kalverra/crazed-nft-fans)](https://goreportcard.com/report/github.com/kalverra/crazed-nft-fans)
[![codecov](https://codecov.io/gh/kalverra/crazed-nft-fans/branch/main/graph/badge.svg)](https://codecov.io/gh/kalverra/crazed-nft-fans)
[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/kalverra/crazed-nft-fans/main/LICENSE)

A new NFT has dropped, but its location is a secret! Fans have become crazed, and are sending funds to random wallets and calling gas guzzling contracts as fast as they can to hopefully snag it.

Using the fans on a simulated network (like [geth in dev mode](https://geth.ethereum.org/docs/getting-started/dev-mode)) can help you emulate network congestion events.

If using a simulated geth instance, you might find it helpful to [turn on metrics](https://geth.ethereum.org/docs/interface/metrics).

## Test

`make test`

to run basic tests in standard go format, or

`make test_fancy`

to run basic tests with a prettier output, or

`make test_integration`

to launch a simulated geth node to run all possible tests.

### Can I use this to cause chaos on ethereum mainnet or testnets?

Maybe? But I wouldn't recommend it, unless you are looking for a way to become very poor, very fast. There's far more fun ways to do that anyway.
