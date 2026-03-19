.PHONY: build run fmt test vet validate-desktop print-version check-release-tag release release-snapshot aur-srcinfo aur-update clean

AUTHOR ?= miniguys
BINARY ?= desktopify-lite
DIST_DIR ?= dist
APP_PKG := github.com/miniguys/desktopify-lite/internal/app
TAG ?= $(shell git describe --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true)
VERSION ?= $(if $(TAG),$(patsubst v%,%,$(TAG)),$(shell ./scripts/current-version.sh))
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(APP_PKG).version=$(VERSION) -X $(APP_PKG).commit=$(COMMIT) -X $(APP_PKG).buildDate=$(BUILD_DATE) -X $(APP_PKG).author=$(AUTHOR)
GOFILES := $(shell find . -name '*.go' -not -path './.git/*')

print-version:
	@echo $(VERSION)

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run -trimpath -ldflags "$(LDFLAGS)" .

fmt:
	gofmt -w $(GOFILES)

test:
	go test ./...

vet:
	go vet ./...

validate-desktop:
	./scripts/validate-generated-desktop.sh

check-release-tag:
	@test -n "$(TAG)" || (echo "release target requires an exact v* tag on HEAD" >&2; exit 1)

release: check-release-tag clean
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-arm64 .
	sha256sum $(DIST_DIR)/$(BINARY)-linux-amd64 $(DIST_DIR)/$(BINARY)-linux-arm64 > $(DIST_DIR)/checksums.txt

release-snapshot:
	goreleaser release --snapshot --clean

aur-srcinfo:
	cd packaging/aur/desktopify-lite && makepkg --printsrcinfo > .SRCINFO

aur-update:
	./scripts/update-aur-metadata.sh $(VERSION)

clean:
	rm -rf $(DIST_DIR) $(BINARY) \
		packaging/aur/desktopify-lite/pkg \
		packaging/aur/desktopify-lite/src
