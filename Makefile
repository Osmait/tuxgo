BINARY   := tuxgo
PKG      := ./cmd/tuxgo
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install test clean lint fmt vet

## build: Compile the binary
build:
	go build $(LDFLAGS) -o $(BINARY) $(PKG)

## install: Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) $(PKG)

## test: Run all tests
test:
	go test -v ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	go clean -cache

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format all Go files
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
