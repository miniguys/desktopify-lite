#!/usr/bin/env bash
set -euo pipefail

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

export XDG_DATA_HOME="$workdir/xdg-data"
export XDG_CONFIG_HOME="$workdir/xdg-config"

name='CI Example App'
url='https://example.com/app?a=1&b=2'
browser_bin="${BROWSER_BIN:-chromium}"
url_template="${URL_TEMPLATE:---app={url}}"
extra_flags="${EXTRA_FLAGS:---profile-directory=Default}"

# This is a launcher-generation smoke test, not browser execution.
# CI may override BROWSER_BIN / URL_TEMPLATE / EXTRA_FLAGS to cover a few
# representative Chromium-style launcher shapes without hardcoding one case.
GOFLAGS='' go run . \
  --url="$url" \
  --name="$name" \
  --skip-icon \
  --browser="$browser_bin" \
  --url-template="$url_template" \
  --extra-flags="$extra_flags"

launcher="$XDG_DATA_HOME/applications/CI_Example_App.desktop"

if [[ ! -f "$launcher" ]]; then
  echo "generated desktop file not found: $launcher" >&2
  exit 1
fi

desktop-file-validate "$launcher"
