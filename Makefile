BIN_DIR = bin
export GOPATH ?= $(shell go env GOPATH)
export GO111MODULE ?= on
current_dir = $(shell pwd)

lint:
	${BIN_DIR}/golangci-lint --color=always run ./... -v

golangci:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${BIN_DIR} v1.42.0

go_mod:
	go mod download

test:
	go test -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts)

test_fancy:
	go install github.com/haveyoudebuggedit/gotestfmt/v2/cmd/gotestfmt@latest	
	set -euo pipefail

	go test -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts) 2>&1 | tee /tmp/gotest.log | gotestfmt

test_race:
	go install github.com/haveyoudebuggedit/gotestfmt/v2/cmd/gotestfmt@latest	
	set -euo pipefail
	
	go test -race -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts) 2>&1 | tee /tmp/gotest.log | gotestfmt

test_integration: clean_test_node start_test_node
	go install github.com/haveyoudebuggedit/gotestfmt/v2/cmd/gotestfmt@latest	
	set -euo pipefail

	go test -race -tags integration -count=1 -json -v -coverprofile=profile.cov $(shell go list ./... | grep -v /contracts) 2>&1 | tee /tmp/gotest.log | gotestfmt
	-docker rm --force test_geth

start_test_node:
	docker build -t kalverra/test_geth -f Dockerfile.test_geth .
	docker run --name test_geth -d -p 8545:8545 -p 8546:8546 test_geth

clean_test_node:
	echo "Cleaning Up Test Env"
	-docker rm --force test_geth

start_realistic_node:
	docker build -t kalverra/realistic_geth -f Dockerfile.realistic_geth .
	docker run --rm --name realistic_geth -d -p 8545:8545 -p 8546:8546 realistic_geth

clean_realistic_node:
	-docker rm --force realistic_geth

# Requires the blockscout repo cloned in a nearby folder
start_blockscout:
	export COIN=ETH && \
	export ETHEREUM_JSONRPC_VARIANT=geth && \
	export ETHEREUM_JSONRPC_HTTP_URL=http://host.docker.internal:8545 && \
	export ETHEREUM_JSONRPC_WS_URL=ws://host.docker.internal:8546 && \
	cd ../blockscout/docker && $(MAKE) start

build:
	go build -o crazed-nft-fans ./main
	chmod +x ./crazed-nft-fans

run:
	go run ./main

run_auto:
	go run ./main -autoStart=true

docker_build:
	docker build . -t kalverra/crazed-nft-fans
