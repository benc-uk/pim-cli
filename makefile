ROOT_DIR := $(shell git rev-parse --show-toplevel)
DEV_DIR := $(ROOT_DIR)/.dev
VERSION := $(shell git describe --tags --abbrev=0 --dirty=-dev 2>/dev/null || echo "0.0.0-dev")

.DEFAULT_GOAL := help

.PHONY: help install build build-win build-mac lint tidy clean ver

help: # 💬 Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: # 📦 Install dependencies
	go mod download
	go mod download -modfile=$(DEV_DIR)/tools.mod
	
build: # 🔨 Build the Go binary for Linux
	go build -o bin/pim-cli -ldflags="-X 'main.version=$(VERSION)'" main.go

build-win: # 🔨 Build the Go binary for Windows
	GOOS=windows GOARCH=amd64 go build -o bin/pim-cli.exe -ldflags="-X 'main.version=$(VERSION)'" main.go

build-mac: # 🔨 Build the Go binary for macOS
	GOOS=darwin GOARCH=amd64 go build -o bin/pim-cli-mac -ldflags="-X 'main.version=$(VERSION)'" main.go

lint: # ✨ Run golangci-lint
	go tool -modfile=.dev/tools.mod golangci-lint run --config $(DEV_DIR)/golangci.yaml

tidy: # 🧹 Tidy Go modules
	go mod tidy

clean: # 🧹 Clean build artifacts
	rm -rf bin/

ver: # 🧲 Show the current version
	@echo $(VERSION)