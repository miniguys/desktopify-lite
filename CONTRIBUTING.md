# Contributing

Thanks for helping with `desktopify-lite`.

## Before opening a PR

- open an issue first for behavior changes or larger features
- keep the scope small and focused
- prefer Linux-oriented changes that match the project scope
- update tests when behavior changes
- update `README.md` when flags, config, or user-visible behavior changes

## Development

```bash
make build
go test ./...
go vet ./...
go build ./...
```

For release-oriented verification, a local snapshot build is also useful:

```bash
goreleaser release --snapshot --clean
```

## Style

- keep dependencies minimal
- prefer straightforward stdlib-first solutions
- keep CLI behavior predictable for interactive and non-interactive modes
- avoid expanding scope beyond generating Linux `.desktop` launchers

## Pull requests

Please include:

- what changed
- why it changed
- any user-visible behavior changes
- docs or tests updated if needed


## Project map

- `main.go`, `run_input.go`, `ui.go` - CLI flow and interactive prompts
- `config*.go` - config loading, defaults, and config commands
- `desktop.go` - `.desktop` entry generation and writing
- `icon.go` - icon download planning, HTML/manifest parsing, validation, and saving
- `fileutil.go` - filesystem helpers and atomic writes
- `runtime.go` - runtime options and path resolution

## Scope guardrails

In scope:

- Linux launcher generation
- fixes that improve predictable CLI behavior
- tests for icon parsing, config precedence, and desktop entry rendering

Out of scope unless discussed first:

- non-Linux platform support
- bundling browsers or creating isolated runtimes
- turning the project into a general app packaging system

## Testing icon logic locally

The icon pipeline is intentionally strict about content validation. When changing `icon.go` or related tests:

- serve real image bytes in HTTP-based tests; do not fake PNG or JPEG bodies with plain strings
- cover both direct icon URLs and HTML/manifest discovery paths when behavior changes
- keep refresh/overwrite cases tested so existing launcher names update their icon files correctly
