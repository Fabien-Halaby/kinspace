BINARY_NAME := kinspace
BINARY_PATH := backend/bin/$(BINARY_NAME)
BACKEND_DIR := backend

.PHONY: all build test test-integration lint fmt tidy bin run mobile-dev clean help

all: tidy fmt lint build test

## Build the Go backend (compile check for every package)
build:
	cd $(BACKEND_DIR) && go build ./...

## Run backend unit tests
test:
	cd $(BACKEND_DIR) && go test -race -count=1 ./...

## Run backend integration tests against a Postgres database
## (requires TEST_DATABASE_URL to point at a scratch database)
test-integration:
	cd $(BACKEND_DIR) && go test -race -count=1 -tags integration ./internal/storage/postgres/

## Lint the backend with golangci-lint
lint:
	cd $(BACKEND_DIR) && go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./...

## Format backend sources
fmt:
	cd $(BACKEND_DIR) && gofmt -w .

## Tidy backend modules
tidy:
	cd $(BACKEND_DIR) && go mod tidy

## Compile the API executable to backend/bin/kinspace
bin:
	mkdir -p $(BACKEND_DIR)/bin
	cd $(BACKEND_DIR) && go build -o bin/$(BINARY_NAME) ./cmd/api

## Run the API server
run:
	cd $(BACKEND_DIR) && go run ./cmd/api/main.go

## Start the Expo / React Native mobile app
mobile-dev:
	cd mobile && npx expo start

## Remove build artifacts and test cache
clean:
	rm -rf $(BACKEND_DIR)/bin
	cd $(BACKEND_DIR) && go clean -cache -testcache ./...

help:
	@echo "Available Makefile commands:"
	@echo ""
	@echo "  make build             Compile-check all backend packages"
	@echo "  make test              Run backend unit tests"
	@echo "  make test-integration  Run backend integration tests (needs TEST_DATABASE_URL)"
	@echo "  make lint              Run golangci-lint on the backend"
	@echo "  make fmt               Format backend sources"
	@echo "  make tidy              Tidy backend Go modules"
	@echo "  make bin               Build the API executable"
	@echo "  make run               Start the Go API server"
	@echo "  make mobile-dev        Start the Expo / React Native mobile app"
	@echo "  make clean             Remove build artifacts and test cache"
