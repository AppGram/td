# td Makefile

BINARY_NAME=td
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: all build build-signed clean install test lint

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

build-signed: build
	@if [ "$$(uname)" = "Darwin" ] && [ -n "$(APPLE_DEVELOPER_ID)" ]; then \
		echo "Signing binary..."; \
		codesign --force --options runtime --sign "$(APPLE_DEVELOPER_ID)" $(BINARY_NAME); \
		echo "Binary signed successfully"; \
	elif [ "$$(uname)" = "Darwin" ]; then \
		echo "Skipping code signing (set APPLE_DEVELOPER_ID to enable)"; \
	else \
		echo "Skipping code signing (not macOS)"; \
	fi

install: build
	@echo "Installing to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "Installed to ~/.local/bin/$(BINARY_NAME)"

uninstall:
	@rm -f ~/.local/bin/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

test:
	go test -v ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

clean:
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@echo "Cleaned"

# Cross-compilation targets
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

build-all: clean
	@mkdir -p dist
	@echo "Building for all platforms..."
	$(MAKE) build-linux
	$(MAKE) build-darwin
	$(MAKE) build-windows
	@echo "Done! Binaries in dist/"

# Release with goreleaser
release:
	goreleaser release --clean

release-snapshot:
	goreleaser release --snapshot --clean

# Development
dev: build
	./$(BINARY_NAME)

run: build
	./$(BINARY_NAME)
