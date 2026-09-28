BINARY_NAME := vertex-on-edge
BIN_DIR := bin
CMD_DIR := ./cmd/vertex-on-edge

VERSION := $(shell grep 'Version.*=' internal/config/config.go | grep -v '//' | head -1 | cut -d'"' -f2)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-s -w \
	-X github.com/yashchokhat/vertex-on-edge/internal/config.Version=$(VERSION) \
	-X github.com/yashchokhat/vertex-on-edge/internal/config.Commit=$(COMMIT) \
	-X github.com/yashchokhat/vertex-on-edge/internal/config.BuildDate=$(BUILD_DATE)"

.PHONY: build run test lint clean fmt vet

build:
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

run: build
	./$(BIN_DIR)/$(BINARY_NAME)

test:
	go test -v -race ./...

lint: fmt vet
	@echo "Lint complete."

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
	go clean
	@echo "Cleaned."
