# AUR packaging

This repository ships AUR-ready metadata in `packaging/aur/desktopify-lite/`.

## Stable package

The included `PKGBUILD` builds from the exact Git tag `v1.0.0`, so future release bumps only need:

1. `./scripts/update-aur-metadata.sh 1.0.1`
2. copy `packaging/aur/desktopify-lite/PKGBUILD` and `.SRCINFO` into the root of your cloned AUR repo
3. commit and push the AUR repo

## Why the source uses a Git tag

This keeps the AUR metadata simple and avoids hand-editing source checksums for every release.
The package is still pinned to an exact release tag rather than tracking the moving default branch.
