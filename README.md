![Desktopify Lite Preview](assets/preview.gif)
# desktopify-lite

Stable CLI tool that generates a Linux `.desktop` launcher for a website.

It launches an existing Chromium-style browser in app mode. It does **not** create an isolated runtime or a separate browser profile by itself.

## Why this exists

Creating `.desktop` launchers by hand is easy once, but annoying to repeat across machines (or to keep consistent in dotfiles).
`desktopify-lite` turns that into a single command, with:

- sensible defaults for Chromium-based browsers (`--app={url}`)
- automatic icon discovery (best-effort) with explicit override via `--icon-url`
- reproducible releases (checksums + documented release process)

## Stability

`desktopify-lite` is stable from `1.0.0` onward.

What that means here:

- the CLI and config surface are expected to stay compatible across `1.x`
- release builds are reproducible in a standard Go toolchain without pulling UI dependencies at build time
- release archives include checksums, and the repository includes a documented release process

See `SUPPORT.md` for the support policy and `RELEASE.md` for the release checklist.

## Install

```bash
go install github.com/miniguys/desktopify-lite@latest
export PATH=$PATH:$(go env GOPATH)/bin # if you don't have ~/go/bin in PATH
```

Or build locally:

```bash
make build
```

Create release binaries:

```bash
make release
```


## Quick start

Create a launcher in one command:

```bash
desktopify-lite --url='https://example.com' --name='Example'
```

This writes a `.desktop` file to:

- `$XDG_DATA_HOME/applications/Example.desktop`
- or `~/.local/share/applications/Example.desktop` if `XDG_DATA_HOME` is not set

Then refresh app launchers (varies by desktop environment) or log out/in, and search for **Example** in your app menu.

Show the binary version:

Tagged installs done via `go install github.com/miniguys/desktopify-lite@latest` report the module version even when linker metadata is not injected.

```bash
desktopify-lite --version
desktopify-lite version
```

## What it creates

- `$XDG_DATA_HOME/applications/<name>.desktop` (falls back to `~/.local/share/applications/<name>.desktop`)
- `$XDG_DATA_HOME/icons/<name>.(svg|png|jpg|jpeg|webp|ico)` when an icon is found (falls back to `~/.local/share/icons/...`)

If icon resolution fails during auto-discovery, the launcher is still created.
If `--icon-url` is passed explicitly and that icon cannot be fetched, copied, or parsed, the command exits with an error.

Generated launchers include `TryExec=<browser>` when a browser binary is known.

## Interactive mode

```bash
desktopify-lite
```

Interactive prompts cover:

- website URL
- icon URL
- launcher name
- browser binary
- URL template
- extra flags
- `StartupWMClass`
- proxy URL

## Non-interactive mode

Useful for scripts, dotfiles, setup repos, devcontainers, and automation.

```bash
desktopify-lite \
  --url='https://example.com?a=1&b=2' \
  --name='Example App' \
  --icon-url='./icon.png' \
  --browser=chromium \
  --url-template='--app={url}' \
  --extra-flags='--profile-directory=Default' \
  --startup-wm-class='Example App'
```

Supported flags:

- `--url`
- `--name`
- `--icon-url`
- `--skip-icon` / `--no-icon`
- `--browser`
- `--url-template`
- `--extra-flags`
- `--startup-wm-class`
- `--proxy`
- `--version` / `-v`

In non-interactive mode, `--url` and `--name` are required.

`--icon-url` accepts:

- `http://` / `https://` icon URLs
- `file:///absolute/path/icon.svg`
- local file paths such as `./icon.png`, `../assets/icon.svg`, `/opt/icons/app.png`, or `~/icons/app.svg`

Use `--skip-icon` (or `--no-icon`) to disable icon resolution entirely.

## Browser profile model

By default this tool launches your existing browser binary with the selected URL template, for example `--app={url}`.

`url-template` is parsed as a command-line fragment after `{url}` substitution, so it may expand to one or more argv parts.

Examples:

- `--url-template='{url}'` -> passes the URL as a standalone argument
- `--url-template='--app={url}'` -> passes one browser flag
- `--url-template='--new-window --app="{url}"'` -> passes two browser arguments

That means:

- no separate embedded runtime
- no guaranteed profile isolation
- browser cookies, session state, extensions, and policies depend on the browser/profile you launch
- if you need isolation, pass your own browser flags such as a dedicated profile directory

## Config

Config lookup order:

1. `config` next to the compiled binary
2. `$XDG_CONFIG_HOME/miniguys/desktopify-lite/config` (falls back to `~/.config/miniguys/desktopify-lite/config`)

If no config exists, a default config is created in the XDG config location. A `config.example` file is also created there.

Example:

```ini
default_browser=chromium
default_url_template=--app={url}
default_extra_flags=
default_proxy=
# disable_google_favicon=true
# with_debug=true
```

## Supported platforms

- Linux desktop environments that support `.desktop` launchers
- Chromium-style browsers such as Chromium, Chrome, Thorium, Brave, Edge, and Vivaldi

## Validation coverage

Generated launchers are validated in CI with `desktop-file-validate`.

CI covers:

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `go build -trimpath ./...`
- shell syntax checks for `scripts/*.sh`
- generated `.desktop` validation across representative Chromium-style launch shapes

## Release integrity

Release archives include `checksums.txt`.

```bash
./scripts/verify-release-checksums.sh
```

## Known limitations

- Linux-only utility; it does not create native launchers for macOS or Windows
- browser profile isolation depends on the browser flags you pass
- icon discovery depends on what the target website exposes via direct image URLs, HTML `<link rel=...>` tags, manifest icons, or favicon fallbacks
- the generated launcher does not force a desktop database refresh; on some systems you may need to run `update-desktop-database "$XDG_DATA_HOME/applications"`

## AUR packaging

AUR metadata for the stable package is included under `packaging/aur/desktopify-lite/`.

To refresh it for the next release:

```bash
./scripts/update-aur-metadata.sh 1.0.1
```

Then copy `PKGBUILD` and `.SRCINFO` from that directory into the root of your cloned AUR repository and push the AUR repo.
