# Makefile for qodercli-reverse (reconstructed Qoder CLI)
# Usage:
#   make build      - build production binary into ./bin
#   make install    - build and install to ~/.local/bin (macOS/Linux)
#   make test       - run go vet + go test
#   make clean      - remove build artifacts
#   make release    - build release tarball for current platform

GO        ?= go
MODULE    := github.com/alingse/qodercli-reverse/decompiled
PKGDIR    := decompiled
BINDIR    := bin
BINARY    := qodercli-reverse
VERSION   ?= v0.1.30
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME:= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -X $(MODULE)/version.Version=$(VERSION) \
             -X main.commit=$(COMMIT) \
             -X main.buildTime=$(BUILD_TIME)

UNAME_S   := $(shell uname -s)
UNAME_M   := $(shell uname -m)
ifeq ($(UNAME_S),Darwin)
  ifeq ($(UNAME_M),arm64)
    GOOS   := darwin
    GOARCH := arm64
  else
    GOOS   := darwin
    GOARCH := amd64
  endif
else
  ifeq ($(UNAME_M),aarch64)
    GOOS   := linux
    GOARCH := arm64
  else
    GOOS   := linux
    GOARCH := amd64
  endif
endif

PREFIX    ?= $(HOME)/.local
INSTALL_BIN := $(PREFIX)/bin/$(BINARY)

.PHONY: all build install test clean release vet

all: build

build:
	@mkdir -p $(BINDIR)
	cd $(PKGDIR) && GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS)" -o ../$(BINDIR)/$(BINARY) .
	@echo "Built $(BINDIR)/$(BINARY) ($(GOOS)/$(GOARCH))"

vet:
	cd $(PKGDIR) && $(GO) vet ./...

test: vet
	cd $(PKGDIR) && $(GO) test ./...

install: build
	@mkdir -p $(PREFIX)/bin
	cp $(BINDIR)/$(BINARY) $(INSTALL_BIN)
	chmod +x $(INSTALL_BIN)
	@echo "Installed to $(INSTALL_BIN)"
	@echo "Ensure $(PREFIX)/bin is on your PATH: export PATH=\"$(PREFIX)/bin:\$$PATH\""

clean:
	rm -rf $(BINDIR) /tmp/$(BINARY)-*

release: build
	@mkdir -p $(BINDIR)/release
	cd $(BINDIR) && tar czf release/$(BINARY)_$(GOOS)_$(GOARCH).tar.gz $(BINARY)
	@echo "Release tarball: $(BINDIR)/release/$(BINARY)_$(GOOS)_$(GOARCH).tar.gz"

doctor: build
	$(BINDIR)/$(BINARY) doctor
