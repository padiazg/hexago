MODULE   := $(shell head -1 go.mod | awk '{print $$2}')
BINARY   := $(notdir $(MODULE))
VERSION_PKG := $(MODULE)/internal/version

LDFLAGS  := -X $(VERSION_PKG).version=$(shell git describe --tags --always --dirty)
LDFLAGS  += -X $(VERSION_PKG).commit=$(shell git rev-parse HEAD)
LDFLAGS  += -X $(VERSION_PKG).buildDate=$(shell date -Iseconds)

.PHONY: build test lint clean fmt mod-tidy help install coverage e2e

build:
	@echo "Building $(BINARY)..."
	@go build -o $(BINARY) -ldflags "$(LDFLAGS)"

clean:
	rm -f $(BINARY) coverage.out mutation*.json

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

mod-tidy:
	go mod tidy

install: build
	cp $(BINARY) $(GOPATH)/bin/

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

preflight: vet lint
	@gremlins unleash --timeout-coefficient 20 -S "l" --integration --output=mutation.json && \
	go-crap scan --exclude ".*_test.go" --top 10 --mutation-report mutation.json

e2e:
	@echo "Running hexago E2E tests..."
	@bash scripts/e2e.sh

help:
	@echo "build     - compile binary"
	@echo "test      - run tests with race detector"
	@echo "lint      - run golangci-lint"
	@echo "clean     - remove build artifacts"
	@echo "fmt       - format source code"
	@echo "mod-tidy  - tidy go.mod"
	@echo "install   - build and copy to GOPATH/bin"
	@echo "coverage  - generate coverage report"
	@echo "e2e       - run full generator E2E tests"
	@echo "help      - show this help"
