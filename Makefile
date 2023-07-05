BIN_DIR = bin
export GOPATH ?= $(shell go env GOPATH)
export GO111MODULE ?= on
current_dir = $(shell pwd)

lint:
	golangci-lint --color=always run ./... --fix

golangci:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${BIN_DIR} v1.42.0

go_mod:
	go mod download

test:
	go test -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts)

install_gotestfmt:
	go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest	
	set -euo pipefail

test_fancy: install_gotestfmt
	go test -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts) 2>&1 | tee /tmp/gotest.log | gotestfmt

test_race: install_gotestfmt
	go test -race -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts) 2>&1 | tee /tmp/gotest.log | gotestfmt

test_integration: clean_test_node start_test_node install_gotestfmt
	LOG_LEVEL="error" \
	go test $(args) -timeout 5m -race -tags integration -count=1 -json -coverprofile=profile.cov $(shell go list ./... | grep -v /guzzle) \
	2>&1 | tee /tmp/gotest.log | gotestfmt
	-docker rm --force test-geth
	go tool cover -html=profile.cov 

test_integration_verbose: clean_test_node start_test_node install_gotestfmt
	go test $(args) -timeout 5m -race -tags integration -count=1 -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /guzzle) \
	2>&1 | tee /tmp/gotest.log | gotestfmt
	-docker rm --force test-geth
	go tool cover -html=profile.cov 

start_test_node:
	docker build -t kalverra/test-geth -f Dockerfile.test-geth .
	docker run --name test-geth -d -p 8545:8545 -p 8546:8546 kalverra/test-geth

clean_test_node:
	echo "Cleaning Up Test Env"
	-docker rm --force test-geth

start_realistic_node:
	docker build -t kalverra/realistic-geth -f Dockerfile.realistic-geth .
	docker run --rm --name realistic-geth -d -p 8545:8545 -p 8546:8546 kalverra/realistic-geth

clean_realistic_node:
	-docker rm --force realistic-geth

# Requires the blockscout repo cloned in a nearby folder
start_blockscout:
	export COIN=ETH && \
	export ETHEREUM_JSONRPC_VARIANT=geth && \
	export ETHEREUM_JSONRPC_HTTP_URL=http://host.docker.internal:8545 && \
	export ETHEREUM_JSONRPC_WS_URL=ws://host.docker.internal:8546 && \
	cd ../blockscout/docker && $(MAKE) start

start_blockscout_compose:
	cd ../blockscout/docker-compose && \
	docker-compose -f docker-compose-no-build-geth.yml up -d
	echo "Find blockscout at http://localhost:4000"

stop_blockscout_compose:
	cd ../blockscout/docker-compose && \
	docker-compose -d -f docker-compose-no-build-geth.yml down

build:
	go build -o crazed-nft-fans ./main
	chmod +x ./crazed-nft-fans

run:
	go run ./main

docker_build:
	docker build . -t kalverra/crazed-nft-fans
