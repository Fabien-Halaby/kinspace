BINARY_NAME=kinspace
BUILD_DIR=bin
BACKEND_DIR=backend

.PHONY: all build test bin run mobile-dev clean help

all: test build

build:
	cd $(BACKEND_DIR) && go build -v ./...

test:
	cd $(BACKEND_DIR) && go test -v ./...

bin:
	mkdir -p $(BACKEND_DIR)/$(BUILD_DIR)
	cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api/main.go

run:
	cd $(BACKEND_DIR) && go run ./cmd/api/main.go

mobile-dev:
	cd mobile && npx expo start

clean:
	rm -rf $(BACKEND_DIR)/$(BUILD_DIR)
	cd $(BACKEND_DIR) && go clean -cache -testcache ./...

help:
	@echo "Available Makefile commands:"
	@echo ""
	@echo "  make build       Check compilation for all Go backend packages"
	@echo "  make test        Run all unit tests"
	@echo "  make bin         Compile executable binary to backend/bin/kinspace"
	@echo "  make run         Start the Go API server"
	@echo "  make mobile-dev  Start the Expo / React Native mobile app"
	@echo "  make clean       Remove compiled binaries and clean test cache"
	@echo "  make help        Show this help message"