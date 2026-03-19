#!/usr/bin/env bash
set -euo pipefail

if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  raw="$(git describe --tags --dirty --always --match 'v[0-9]*' 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || true)"
  raw="${raw#v}"
  if [[ -n "$raw" ]]; then
    printf '%s\n' "$raw"
    exit 0
  fi
fi

printf 'dev\n'
