.PHONY: build

build: pkg=github.com/padiazg/hexago/pkg/version
build: ldflags = -X $(pkg).version=$(shell git describe --tags --always --dirty) 
build: ldflags += -X $(pkg).commit=$(shell git rev-parse HEAD)
build: ldflags += -X $(pkg).buildDate=$(shell date -Iseconds)

build:
	@echo "Building hexago..."
	@go build -o hexago -ldflags "$(ldflags)"
