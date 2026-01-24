ROOT_DIR := $(shell git rev-parse --show-toplevel)
DEV_DIR := $(ROOT_DIR)/.dev
PACKAGE := github.com/benc-uk/pimg-cli
VERSION := $(shell git describe --tags --abbrev=0 --dirty=-dev 2>/dev/null || echo "0.0.0-dev")

.DEFAULT_GOAL := help

.PHONY: help build test run watch lint tidy install story web watch-web serve clean ver test-czech

help: # 💬 Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: # 🔨 Build the Go binary
	go build -o bin/pimg -ldflags="-X 'main.version=$(VERSION)'" main.go

install: # 📦 Install dependencies
	go mod download
	go mod download -modfile=$(DEV_DIR)/tools.mod
	
# build-win: # 🔨 Build the Go binary for Windows
# 	GOOS=windows GOARCH=amd64 go build -o bin/gozm.exe -ldflags="-X 'main.version=$(VERSION)'" $(PACKAGE)/impl/terminal

# build-mac: # 🔨 Build the Go binary for macOS
# 	GOOS=darwin GOARCH=amd64 go build -o bin/mac-gozm -ldflags="-X 'main.version=$(VERSION)'" $(PACKAGE)/impl/terminal

# test: # 🧪 Run tests
# 	go test -v ./...

# test-czech: build # 🔬 Run tests with Czech test suite
# 	clear
# 	./bin/gozm -file=stories/czech.z3 -debug=$(DEBUG)

# run: # 🚀 Run the terminal app
# 	clear
# 	go run $(PACKAGE)/impl/terminal -file=test/$(STORY).z3 -debug=$(DEBUG)

# watch: # 👀 Watch for changes and run the terminal app
# 	clear
# 	go tool -modfile=.dev/tools.mod air -c $(DEV_DIR)/air.toml

# lint: # ✨ Run golangci-lint
# 	go tool -modfile=.dev/tools.mod golangci-lint run --config $(DEV_DIR)/golangci.yaml

# tidy: # 🧹 Tidy Go modules
# 	go mod tidy
# 	go mod tidy -modfile=$(DEV_DIR)/tools.mod

# install: # 📦 Install dependencies
# 	go mod download
# 	go mod download -modfile=$(DEV_DIR)/tools.mod

# story: # 📚 Compile and dump the story file
# 	inform6 -v3 ./test/$(STORY).inf ./test/$(STORY).z3
# 	./tools/unz ./test/$(STORY).z3 > ./test/$(STORY).dump.txt

# web: # 🔨 Build the web app
# 	rm -f web/main.wasm 
# 	echo "export const version = \"$(VERSION)\";" > web/js/version.js
# 	GOOS=js GOARCH=wasm go build -o web/main.wasm $(PACKAGE)/impl/web

# clean: # 🧹 Clean build artifacts
# 	rm -rf bin/
# 	rm -f web/main.wasm 

# watch-web: # 👀 Watch for changes and build the web app
# 	clear
# 	go tool -modfile=.dev/tools.mod air -c $(DEV_DIR)/air-wasm.toml

# serve: web # 🌐 Serve the web app
# 	npx vite web/

# ver: # 🧲 Show the current version
# 	@echo $(VERSION)