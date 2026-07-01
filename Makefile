# Gatekeeper — Build targets for Linux, macOS, and Windows

BINARY_NAME  := gatekeeper
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE         := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS      := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

BUILD_DIR    := dist
GOFLAGS      := -buildvcs=false

# Current platform
GOOS_CURRENT := $(shell go env GOOS)
GOARCH_CURRENT := $(shell go env GOARCH)

.PHONY: all build clean test vet lint

all: build

## build — compile for the current platform
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS_CURRENT)-$(GOARCH_CURRENT) .
	@echo "built $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS_CURRENT)-$(GOARCH_CURRENT)"

## linux — compile for Linux (amd64 + arm64)
linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "built linux-amd64 linux-arm64"

## macos — compile for macOS (amd64 + arm64)
macos:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "built darwin-amd64 darwin-arm64"

## windows — compile for Windows (amd64)
windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "built windows-amd64.exe"

## release — compile for all platforms
release: linux macos windows
	@echo "all platforms built in $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

## clean — remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## test — run all tests
test:
	go test $(GOFLAGS) ./... -count=1

## vet — run go vet
vet:
	go vet ./...

## lint — vet + test
lint: vet test
