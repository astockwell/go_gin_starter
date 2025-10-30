GO ?= go
GO_VERSION := $(shell awk '/^go / {print $$2}' go.mod)
BINARY_NAME ?= go-gin-starter
BIN_DIR ?= bin

.PHONY: help
# Show available make targets and their descriptions
help:
	@awk '/^\.PHONY: / { target = $$2; getline; if ($$0 ~ /^#/) { desc = $$0; sub(/^# */, "", desc); printf "%-18s %s\n", target, desc } }' $(MAKEFILE_LIST)

.PHONY: build
# Compile the project binary for the current platform
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BINARY_NAME) ./...

.PHONY: run
# Build and execute the project locally
run: build
	./$(BIN_DIR)/$(BINARY_NAME)

.PHONY: clean
# Remove build artifacts
clean:
	rm -rf $(BIN_DIR)

.PHONY: fmt
# Format Go sources
fmt:
	$(GO) fmt ./...

.PHONY: vet
# Run go vet static analysis
vet:
	$(GO) vet ./...

.PHONY: test
# Run the test suite
test:
	$(GO) test ./...

.PHONY: tidy
# Synchronize module dependencies
tidy:
	$(GO) mod tidy

.PHONY: deps
# Update dependencies and verify modules
deps:
	$(GO) get -u ./...
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: docker-build
# Build the Docker image using the module's Go version
docker-build:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(BINARY_NAME):latest .

.PHONY: docker-run
# Build (if needed) and run the Docker image
docker-run: docker-build
	docker run --rm -p 8080:8080 $(BINARY_NAME):latest

.PHONY: print-go-version
# Print the Go toolchain version from go.mod
print-go-version:
	@echo $(GO_VERSION)

.PHONY: cross-compile
# Build binaries for linux/amd64, linux/arm64, darwin/amd64, and darwin/arm64
cross-compile:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 BIN_DIR=$(BIN_DIR) BINARY_NAME=$(BINARY_NAME)-linux-amd64 $(MAKE) build
	GOOS=linux GOARCH=arm64 BIN_DIR=$(BIN_DIR) BINARY_NAME=$(BINARY_NAME)-linux-arm64 $(MAKE) build
	GOOS=darwin GOARCH=amd64 BIN_DIR=$(BIN_DIR) BINARY_NAME=$(BINARY_NAME)-darwin-amd64 $(MAKE) build
	GOOS=darwin GOARCH=arm64 BIN_DIR=$(BIN_DIR) BINARY_NAME=$(BINARY_NAME)-darwin-arm64 $(MAKE) build
