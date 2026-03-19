#!/usr/bin/env bash
set -euo pipefail

version="${1:-$(./scripts/current-version.sh)}"
commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
author="miniguys"
out_dir="dist/pre-release-check"
bin="$out_dir/desktopify-lite"
pkg='github.com/miniguys/desktopify-lite/internal/app'
ldflags="-s -w -X ${pkg}.version=${version} -X ${pkg}.commit=${commit} -X ${pkg}.buildDate=${build_date} -X ${pkg}.author=${author}"

mkdir -p "$out_dir"

echo "==> format check"
unformatted="$(find . -name '*.go' -not -path './.git/*' -print0 | xargs -0 gofmt -l)"
if [[ -n "$unformatted" ]]; then
  echo "$unformatted" >&2
  exit 1
fi

echo "==> shell syntax check"
bash -n scripts/*.sh

echo "==> build local verification binary"
CGO_ENABLED=0 go build -trimpath -ldflags "$ldflags" -o "$bin" .

echo "==> verify injected metadata strings exist"
strings "$bin" | grep -F "$version" >/dev/null
strings "$bin" | grep -F "$commit" >/dev/null
strings "$bin" | grep -F "$build_date" >/dev/null

echo "==> verify version command"
"$bin" --version | grep -F "desktopify-lite $version" >/dev/null

if command -v desktop-file-validate >/dev/null 2>&1; then
  echo "==> validate generated .desktop file"
  ./scripts/validate-generated-desktop.sh
else
  echo "==> skip desktop-file-validate check (desktop-file-validate not installed)"
fi

if [[ "${SKIP_TESTS:-0}" != "1" ]]; then
  echo "==> run unit tests"
  go test ./...

  echo "==> run vet"
  go vet ./...
else
  echo "==> skip go test / go vet (SKIP_TESTS=1)"
fi

if [[ "${SKIP_GORELEASER:-0}" != "1" ]]; then
  command -v goreleaser >/dev/null 2>&1 || {
    echo "goreleaser is required for the snapshot check (or set SKIP_GORELEASER=1)" >&2
    exit 1
  }

  echo "==> build local goreleaser snapshot"
  goreleaser release --snapshot --clean

  echo "==> verify release checksums"
  ./scripts/verify-release-checksums.sh
else
  echo "==> skip goreleaser snapshot (SKIP_GORELEASER=1)"
fi
