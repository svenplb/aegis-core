.PHONY: build build-scan build-tui test lint clean

BINARY_SCAN := bin/aegis-scan
BINARY_TUI  := bin/aegis

build: build-scan build-tui

build-scan:
	go build -o $(BINARY_SCAN) ./cmd/aegis-scan

build-tui:
	go build -o $(BINARY_TUI) ./cmd/aegis

test:
	go test ./... -v

test-race:
	go test ./... -race -v

bench:
	go test ./internal/scanner/ -bench=. -benchmem

lint:
	go vet ./...

clean:
	rm -rf bin/
