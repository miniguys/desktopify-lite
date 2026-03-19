#!/usr/bin/env bash
set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
  version="$(git describe --tags --exact-match --match 'v[0-9]*' 2>/dev/null || true)"
fi
version="${version#v}"

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "usage: $0 <semver>; got '$version'" >&2
  exit 1
fi

aur_dir="packaging/aur/desktopify-lite"
pkgbuild="$aur_dir/PKGBUILD"
srcinfo="$aur_dir/.SRCINFO"

python3 - "$version" "$pkgbuild" <<'PY'
import re, sys, pathlib
version = sys.argv[1]
path = pathlib.Path(sys.argv[2])
text = path.read_text(encoding="utf-8")
text = re.sub(r'^pkgver=.*$', f'pkgver={version}', text, flags=re.M)
text = re.sub(r'^pkgrel=.*$', 'pkgrel=1', text, flags=re.M)
text = re.sub(r'#tag=v[0-9][^"]*', f'#tag=v{version}', text)
path.write_text(text, encoding="utf-8")
PY

if command -v makepkg >/dev/null 2>&1; then
  (
    cd "$aur_dir"
    makepkg --printsrcinfo > .SRCINFO
  )
  echo "updated $pkgbuild and regenerated $srcinfo"
else
  echo "updated $pkgbuild"
  echo "makepkg not found; regenerate $srcinfo on Arch with: (cd $aur_dir && makepkg --printsrcinfo > .SRCINFO)"
fi
