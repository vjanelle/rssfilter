.PHONY: test ginkgo ginkgo-install tidy update-deps build run

GINKGO_PKG := github.com/onsi/ginkgo/v2/ginkgo
GINKGO_BIN := $(shell go env GOPATH)/bin/ginkgo

# Install the ginkgo CLI if it's not already installed.
# Note: relies on GOPATH/bin being on PATH, or uses the absolute path for execution.
ginkgo-install:
	@if [ ! -x "$(GINKGO_BIN)" ]; then \
		echo "Installing ginkgo CLI to $(GINKGO_BIN)"; \
		go install $(GINKGO_PKG)@latest; \
	else \
		echo "ginkgo CLI already installed at $(GINKGO_BIN)"; \
	fi

# Run tests with the ginkgo test runner.
ginkgo:
	@go run $(GINKGO_PKG)@latest -r

# Standard 'make test'
test: ginkgo

tidy:
	go mod tidy

update-deps:
	go get -u ./...
	go mod tidy

build:
	go build ./...

run:
	go run .
