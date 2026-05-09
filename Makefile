.PHONY: all build sshwap kyc-cli test vet lint clean run-host run-cli docker

BIN_DIR := ./bin
GOFLAGS := -trimpath -ldflags="-s -w"
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
