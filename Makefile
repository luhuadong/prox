GO ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: all build fmt fmt-check test test-shell vet check clean

all: check build

build:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o dist/prox ./cmd/prox

fmt:
	gofmt -w cmd internal

fmt-check:
	test -z "$$(gofmt -l cmd internal)"

test:
	$(GO) test ./...

test-shell:
	bash -n internal/shell/prox.bash
	bash tests/bash_hook_test.sh

vet:
	$(GO) vet ./...

check: fmt-check test vet test-shell

clean:
	rm -f dist/prox
