#!/usr/bin/env bash
set -euo pipefail

checksum_file="${1:-dist/checksums.txt}"
if [[ ! -f "$checksum_file" ]]; then
  echo "checksum file not found: $checksum_file" >&2
  exit 1
fi

(
  cd "$(dirname "$checksum_file")"
  sha256sum -c "$(basename "$checksum_file")"
)
