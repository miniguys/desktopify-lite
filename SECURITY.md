# Security Policy

## Supported versions

Only the latest released `1.x` version is supported with security fixes.
Pre-1.0 builds are unsupported.

## Reporting a vulnerability

Do **not** open a public issue for suspected security problems.

Report it privately through GitHub security reporting if it is enabled for the repository, or through another maintainer contact channel listed on the repository profile.

Include:

- affected version or commit
- reproduction steps
- impact
- any suggested fix or mitigation

## Network surface

`desktopify-lite` can fetch the target page, linked manifest files, discovered icon URLs, and the Google favicon fallback as part of normal user-driven icon discovery. These requests are not sandboxed and run with the current user's network permissions.

Proxy settings are supported through `--proxy` and `default_proxy`.
