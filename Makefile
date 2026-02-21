.PHONY: build build-scan build-tui build-server test lint clean docker-server benchmark-accuracy

BINARY_SCAN   := bin/aegis-scan
BINARY_TUI    := bin/aegis
BINARY_SERVER := bin/aegis-server

build: build-scan build-tui build-server

build-scan:
	go build -o $(BINARY_SCAN) ./cmd/aegis-scan

build-tui:
	go build -o $(BINARY_TUI) ./cmd/aegis

build-server:
	go build -o $(BINARY_SERVER) ./cmd/aegis-server

docker-server:
	docker build -f Dockerfile.server -t aegis-server .

test:
	go test ./... -v

test-race:
	go test ./... -race -v

bench:
	go test ./internal/scanner/ -bench=. -benchmem

benchmark-accuracy:
	go test ./internal/scanner/ -run TestBenchmark -v -count=1

lint:
	go vet ./...

clean:
	rm -rf bin/
