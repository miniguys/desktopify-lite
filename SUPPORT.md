# Support

## Stability policy

`desktopify-lite` 1.x is the stable release line.

Stable in this repository means:

- CLI flags and config keys are expected to remain compatible across 1.x unless a security or correctness issue requires a break
- generated `.desktop` files keep the same basic contract: predictable launcher naming, stable file locations, and explicit browser command construction
- changes that alter user-visible behavior must be documented in `CHANGELOG.md`

## Support window

- latest 1.x release: full support
- older 1.x releases: best effort until users can reasonably upgrade
- pre-1.0 releases: no support guarantees

## Getting help

Open an issue for:

- reproducible launcher generation bugs
- browser argument handling regressions
- desktop environment compatibility issues
- icon discovery regressions
