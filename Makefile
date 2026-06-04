APP := pvectl
VERSION := v$(shell date +%Y%m%d)-$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
GOENV := CGO_ENABLED=0

.PHONY: all build fmt check test clean help

all: fmt check test build

build:
	$(GOENV) go build $(LDFLAGS) -o bin/$(APP) ./main.go

fmt:
	go fmt ./...

check:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin dist coverage.out

help:
	@echo "make build  - build binary"
	@echo "make fmt    - format code"
	@echo "make check  - run go vet"
	@echo "make test   - run tests"
	@echo "make clean  - clean generated artifacts"
