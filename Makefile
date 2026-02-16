BINARY   := tuxgo
PKG      := ./cmd/tuxgo
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-s -w -X github.com/Osmait/tuxgo/internal/cli.Version=$(VERSION)"

.PHONY: build install test clean lint fmt vet release release-dry next-version tag

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

## next-version: Show the next version based on conventional commits
next-version:
	@LAST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
	if [ -z "$$LAST_TAG" ]; then \
		echo "v0.1.0"; \
	else \
		MAJOR=$$(echo $$LAST_TAG | sed 's/^v//' | cut -d. -f1); \
		MINOR=$$(echo $$LAST_TAG | sed 's/^v//' | cut -d. -f2); \
		PATCH=$$(echo $$LAST_TAG | sed 's/^v//' | cut -d. -f3); \
		COMMITS=$$(git log $$LAST_TAG..HEAD --pretty=format:"%s" 2>/dev/null); \
		if echo "$$COMMITS" | grep -qE '^.+!:|BREAKING CHANGE'; then \
			echo "v$$((MAJOR + 1)).0.0"; \
		elif echo "$$COMMITS" | grep -qE '^feat(\(.+\))?:'; then \
			echo "v$$MAJOR.$$((MINOR + 1)).0"; \
		else \
			echo "v$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
		fi; \
	fi

## tag: Create and push a version tag based on conventional commits
tag:
	@NEXT=$$($(MAKE) -s next-version); \
	echo "Creating tag $$NEXT"; \
	git tag -a $$NEXT -m "Release $$NEXT"; \
	git push origin $$NEXT

## release-dry: Run goreleaser in dry-run mode (no publish)
release-dry:
	goreleaser release --snapshot --clean

## release: Tag and push to trigger the GitHub Actions release pipeline
release:
	@$(MAKE) tag
	@echo "Tag pushed. GitHub Actions will build and publish the release."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
