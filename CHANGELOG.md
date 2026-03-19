# Changelog

## Unreleased

- removed the external `replace` directive and moved the lipgloss compatibility shim into the repository so `go install github.com/miniguys/desktopify-lite@latest` works for tagged releases again
- made local builds derive their version from Git tags automatically and restricted `make release` to exact tagged commits
- added AUR packaging files under `packaging/aur/desktopify-lite` plus a helper script to refresh `pkgver` and `.SRCINFO`

## 1.0.0 - 2026-03-03

- marked the project as stable for the `1.x` line with an explicit support policy
- added `TryExec` to generated `.desktop` files for better launcher integration on Linux desktops
- changed stale icon cleanup to happen only after a successful `.desktop` write
- made the build self-contained by replacing the runtime styling dependency with a local compatibility shim
- expanded release hardening with checksum verification, CodeQL, Dependabot, support docs, and a maintainer release checklist
- updated CI to cover `go test`, `go test -race`, `go vet`, trimpath builds, shell syntax checks, and generated `.desktop` validation
- added `--version` / `-v` and a `version` command for release build identification
- upgraded the release workflow to run the same preflight checks as CI before publishing artifacts
- made `go install github.com/miniguys/desktopify-lite@latest` report the module version via Go build info when linker metadata is not injected

## 0.1.1 - 2026-03-02

- fixed linker flag targets for injected build metadata (`version`, `commit`, `buildDate`, `author`)
- added CI validation for a generated `.desktop` file with `desktop-file-validate`
- documented network and icon fetching behavior more explicitly
- widened CI launcher validation to cover multiple representative Chromium-style launch configurations
- added a maintainer pre-release check script for metadata and release artifact inspection
- clarified CI validation coverage in the README and added a small tested-on section
- made the maintainer pre-release check less fragile by verifying embedded metadata directly before the optional manual UI pass

## 0.1.0 - 2026-03-01

- initial public release
- interactive and non-interactive launcher creation
- config file support with CLI updates and reset
- icon discovery via direct image URLs, HTML icon tags, manifest icons, and favicon fallbacks
- XDG data directory support for launcher and icon output
- release metadata injected via linker flags
