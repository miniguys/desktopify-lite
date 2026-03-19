# Release process

## Maintainer checklist

1. update `CHANGELOG.md` and docs that changed
2. run `./scripts/pre-release-check.sh`
3. create and push an annotated tag: `git tag -a vX.Y.Z -m 'vX.Y.Z' && git push origin vX.Y.Z`
4. let the tag-triggered GitHub release workflow publish artifacts
5. verify `checksums.txt` from a clean machine with `./scripts/verify-release-checksums.sh`
6. if you publish to AUR, run `./scripts/update-aur-metadata.sh X.Y.Z` and push `packaging/aur/desktopify-lite/{PKGBUILD,.SRCINFO}` to the AUR repo

## Versioning

This project follows semantic versioning for the stable `1.x` line.

- patch: fixes and compatibility-safe improvements
- minor: backward-compatible feature additions
- major: intentional breaking changes

## Automation notes

- `make build` and `make run` derive the version from `git describe --tags --dirty --always`
- `make release` only works when `HEAD` is on an exact `v*` tag, so release binaries cannot accidentally be built with a stale manual version
- GoReleaser derives the published version from the Git tag
- the AUR metadata can be refreshed with `./scripts/update-aur-metadata.sh X.Y.Z`
