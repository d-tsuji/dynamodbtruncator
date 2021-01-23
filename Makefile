BIN := dynamodbtruncator
BUILD_LDFLAGS := "-s -w"
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: deps
deps:
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	sh -c '\
      tmpdir=$$(mktemp -d); \
      cd $$tmpdir; \
      go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.34.1; \
      rm -rf $$tmpdir'

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) -trimpath ./cmd/dynamodbtruncator

.PHONY: test
test: deps
	go test -v ./...

.PHONY: test-cover
test-cover: deps
	go test -v ./... -cover -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

.PHONY: lint
lint: devel-deps
	go vet ./...
	golangci-lint run --config .golangci.yml ./...

.PHONY: clean
clean:
	rm -rf $(BIN)
	go clean


