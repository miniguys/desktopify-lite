![Desktopify Lite Preview](assets/preview.gif)

# desktopify-lite

Turn any website into a desktop app launcher — one command, works with your existing browser.

```bash
desktopify-lite --url='https://notion.so' --name='Notion'
```

That's it. Notion now lives in your app menu with an icon, like any native app.

> Perfect for dotfiles — run once on a new machine and all your web apps are back.

---

## Install

**Arch Linux (AUR)**
```bash
yay -S desktopify-lite
```

**Go install**
```bash
go install github.com/miniguys/desktopify-lite@latest
```

**Build from source**
```bash
make build
```

---

## Why not just write the `.desktop` file by hand?

You can. It takes about 5 minutes once, and 5 minutes again on each new machine, and another 5 when you forget the icon flags syntax. `desktopify-lite` turns that into one command you can put in your dotfiles setup script and never think about again.

What it handles for you:
- automatic icon discovery from the site (favicon, manifest, `<link rel>` tags)
- sensible Chromium app-mode defaults (`--app={url}`)
- XDG-compliant paths for the launcher and icon files
- CI-validated output (`desktop-file-validate` on every release)

---

## Quick examples

**Minimal — URL and name only:**
```bash
desktopify-lite --url='https://linear.app' --name='Linear'
```

**With explicit icon:**
```bash
desktopify-lite --url='https://figma.com' --name='Figma' --icon-url='./figma.png'
```

**Scripted / dotfiles setup:**
```bash
desktopify-lite \
  --url='https://app.slack.com' \
  --name='Slack' \
  --browser=chromium \
  --extra-flags='--profile-directory=Default'
```

**Interactive mode** (prompts for everything):
```bash
desktopify-lite
```

---

## What gets created

```
~/.local/share/applications/Notion.desktop
~/.local/share/icons/Notion.png
```

If icon discovery fails, the launcher is still created without an icon. If you pass `--icon-url` explicitly and it can't be fetched, the command exits with an error.

---

## All flags

| Flag | Description |
|---|---|
| `--url` | Target URL (required in non-interactive mode) |
| `--name` | Launcher name (required in non-interactive mode) |
| `--icon-url` | Icon path or URL. Accepts `https://`, `file://`, or local paths |
| `--skip-icon` / `--no-icon` | Skip icon resolution entirely |
| `--browser` | Browser binary name (e.g. `chromium`, `brave`, `google-chrome`) |
| `--url-template` | How the URL is passed to the browser. Default: `--app={url}` |
| `--extra-flags` | Additional browser flags |
| `--startup-wm-class` | `StartupWMClass` for the `.desktop` file |
| `--proxy` | Proxy URL |
| `--version` / `-v` | Print version |

---

## Config file

On first run, a default config is created at:
```
~/.config/miniguys/desktopify-lite/config
```

```ini
default_browser=chromium
default_url_template=--app={url}
default_extra_flags=
default_proxy=
# disable_google_favicon=true
# with_debug=true
```

Config lookup order: binary directory → XDG config path.

---

## Browser profile model

This tool launches your existing browser binary — it does not create an isolated runtime or a separate profile on its own.

If you need isolation, pass `--extra-flags='--profile-directory=MyProfile'` or similar browser-specific flags.

---

## Supported platforms

- Linux with any desktop environment that supports `.desktop` launchers
- Chromium-style browsers: Chromium, Chrome, Thorium, Brave, Edge, Vivaldi

Not supported: macOS, Windows.

---

## Stability

Stable from `1.0.0` onward. CLI and config surface stay compatible across `1.x`.

Release archives include `checksums.txt`. Verify with:
```bash
./scripts/verify-release-checksums.sh
```

---

## CI coverage

Every release validates:
- `go test ./...` and `go test -race ./...`
- `go vet ./...`
- `go build -trimpath ./...`
- shell syntax checks on `scripts/*.sh`
- generated `.desktop` files with `desktop-file-validate`

---

## AUR packaging

AUR metadata lives under `packaging/aur/desktopify-lite/`. To update for a new release:

```bash
./scripts/update-aur-metadata.sh 1.0.6
```

Then push the updated `PKGBUILD` and `.SRCINFO` to the AUR repo.

---

## Known limitations

- Does not force a desktop database refresh — on some systems run `update-desktop-database ~/.local/share/applications` after install
- Icon discovery depends on what the site exposes (favicon, manifest, `<link rel>` tags)
- No macOS or Windows support
