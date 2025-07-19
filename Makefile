# Makefile for simsdk (NeuroSim Simulation SDK)

# Default Go parameters
GO ?= go
PKG := github.com/neurosimio/simsdk
VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: all build test lint fmt tidy tag clean

## Default: run build
all: build

## Build the SDK (noop if it's a library)
build:
	@echo "Building SDK..."
	@$(GO) build ./...

## Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

## Lint using go vet
lint:
	@echo "Linting..."
	@$(GO) vet ./...

## Format all Go code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

## Ensure go.mod is tidy
tidy:
	@echo "Tidying go.mod..."
	@$(GO) mod tidy

## Create a git tag for release
tag:
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

## Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f *.out *.test