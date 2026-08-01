.PHONY: build test lint clean install

VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -s -w \
            -X github.com/sanketpadhyal/towie/internal/buildinfo.Version=$(VERSION) \
            -X github.com/sanketpadhyal/towie/internal/buildinfo.Commit=$(COMMIT) \
            -X github.com/sanketpadhyal/towie/internal/buildinfo.Date=$(DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/towie ./cmd/towie

test:
	CGO_ENABLED=0 go test -count=1 ./...

test-race:
	go test -race -count=1 ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

install:
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/towie

run: build
	./bin/towie start

tidy:
	go mod tidy
