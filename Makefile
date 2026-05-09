.PHONY: all build sshwap kyc-cli test vet lint clean run-host run-cli docker release-snapshot release-check

BIN_DIR := ./bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
GOFLAGS := -trimpath -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"
ENV     := CGO_ENABLED=0

all: build

build: sshwap kyc-cli

sshwap:
	$(ENV) go build $(GOFLAGS) -o $(BIN_DIR)/sshwap ./cmd/sshwap

kyc-cli:
	$(ENV) go build $(GOFLAGS) -o $(BIN_DIR)/kyc-cli ./cmd/kyc-cli

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null && golangci-lint run || echo "(golangci-lint not installed, skipping)"

clean:
	rm -rf $(BIN_DIR)

run-host: sshwap
	./bin/sshwap -addr 127.0.0.1:23222 -host-key /tmp/sshwap-dev-hostkey

run-cli: kyc-cli
	./bin/kyc-cli

docker:
	docker build -t sshwap:dev .

# Local goreleaser dry-run — produces ./dist/* without publishing.
release-snapshot:
	goreleaser release --snapshot --clean

# Validate .goreleaser.yml without building anything.
release-check:
	goreleaser check
