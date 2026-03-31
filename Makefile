.PHONY: build test fmt lint

BINARY := wc-stats-mcp

build:
	go build -o $(BINARY) ./cmd/wc-stats-mcp

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...
