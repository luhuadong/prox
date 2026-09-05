GO ?= go
GORELEASER ?= goreleaser
INSTALL ?= install
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
DESTDIR ?=

.PHONY: all build install uninstall fmt fmt-check test test-shell test-install vet check snapshot release-check clean

all: check build

build:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o dist/prox ./cmd/prox

install: build
	$(INSTALL) -d "$(DESTDIR)$(BINDIR)"
	$(INSTALL) -m 0755 dist/prox "$(DESTDIR)$(BINDIR)/prox"
	$(INSTALL) -d "$(DESTDIR)$(PREFIX)/share/bash-completion/completions"
	$(INSTALL) -m 0644 internal/shell/completion.bash "$(DESTDIR)$(PREFIX)/share/bash-completion/completions/prox"

uninstall:
	rm -f "$(DESTDIR)$(BINDIR)/prox"
	rm -f "$(DESTDIR)$(PREFIX)/share/bash-completion/completions/prox"

fmt:
	gofmt -w cmd internal

fmt-check:
	test -z "$$(gofmt -l cmd internal)"

test:
	$(GO) test ./...

test-shell:
	bash -n internal/shell/prox.bash
	bash -n internal/shell/completion.bash
	bash tests/bash_hook_test.sh
	bash tests/completion_test.sh

test-install:
	bash tests/install_test.sh

vet:
	$(GO) vet ./...

check: fmt-check test vet test-shell test-install

snapshot:
	$(GORELEASER) release --snapshot --clean

release-check:
	$(GORELEASER) check

clean:
	rm -f dist/prox
