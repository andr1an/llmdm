.PHONY: build test run lint clean build-linux

build:
	go build -o bin/dnd-mcp ./cmd/server

test:
	go test ./... -v -race

run:
	go run ./cmd/server serve

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

# Cross-compile for Linux (for VPS deploy)
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/dnd-mcp-linux ./cmd/server
