.PHONY: build test lint clean install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o codelens-memory ./cmd/codelens-memory

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -f codelens-memory
	rm -rf dist/

install:
	go install -ldflags "-s -w -X main.version=$(VERSION)" ./cmd/codelens-memory

# Cross-compile for all platforms
dist: clean
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/codelens-memory-linux-amd64 ./cmd/codelens-memory
	GOOS=linux   GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/codelens-memory-linux-arm64 ./cmd/codelens-memory
	GOOS=darwin  GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/codelens-memory-darwin-amd64 ./cmd/codelens-memory
	GOOS=darwin  GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/codelens-memory-darwin-arm64 ./cmd/codelens-memory
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/codelens-memory-windows-amd64.exe ./cmd/codelens-memory
