package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	orgConfigDirName      = "miniguys"
	appConfigDirName      = "desktopify-lite"
	localConfigFileName   = "config"
	exampleConfigFileName = "config.example"
)

type ConfigLoadMeta struct {
	ActivePath         string
	CreatedDefaultPath string
	CreatedExamplePath string
}

func DefaultConfig() Config {
	return Config{
		DefaultBrowser:       "chromium",
		DefaultURLTemplate:   "--app={url}",
		DefaultExtraFlags:    "",
		DefaultProxy:         "",
		DisableGoogleFavicon: false,
		WithDebug:            false,
	}
}

func LoadConfig() (Config, ConfigLoadMeta, error) {
	defaultCfg := DefaultConfig()

	binDir, err := executableDir()
	if err != nil {
		return Config{}, ConfigLoadMeta{}, fmt.Errorf("resolve executable dir: %w", err)
	}

	localPath := filepath.Join(binDir, localConfigFileName)

	// 1. Local config next to the binary (portable/dev mode)
	if _, err := os.Stat(localPath); err == nil {
		cfg, err := parseConfigFile(localPath, defaultCfg)
		if err != nil {
			return Config{}, ConfigLoadMeta{ActivePath: localPath}, err
		}
		return cfg, ConfigLoadMeta{ActivePath: localPath}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, ConfigLoadMeta{}, fmt.Errorf("stat local config: %w", err)
	}

	// 2. XDG config (~/.config/miniguys/desktopify-lite/config)
	cfgDir, err := osUserConfigDir()
	if err != nil {
		return Config{}, ConfigLoadMeta{}, fmt.Errorf("resolve user config dir: %w", err)
	}

	xdgDir := filepath.Join(cfgDir, orgConfigDirName, appConfigDirName)
	xdgPath := filepath.Join(xdgDir, localConfigFileName)
	examplePath := filepath.Join(xdgDir, exampleConfigFileName)

	if _, err := os.Stat(xdgPath); err == nil {
		cfg, err := parseConfigFile(xdgPath, defaultCfg)
		if err != nil {
			return Config{}, ConfigLoadMeta{ActivePath: xdgPath}, err
		}
		return cfg, ConfigLoadMeta{ActivePath: xdgPath}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, ConfigLoadMeta{}, fmt.Errorf("stat user config: %w", err)
	}

	// 3. Missing everywhere -> create the config directory structure
	if err := EnsureConfigDir(xdgDir); err != nil {
		return Config{}, ConfigLoadMeta{}, err
	}

	if err := osWriteFileAtomic(xdgPath, []byte(renderConfig(defaultCfg)), 0o644); err != nil {
		return Config{}, ConfigLoadMeta{}, fmt.Errorf("write default config: %w", err)
	}

	// Example config is best-effort and should not fail startup
	_ = osWriteFileAtomic(examplePath, []byte(renderExampleConfig(defaultCfg)), 0o444)

	return defaultCfg, ConfigLoadMeta{
		ActivePath:         xdgPath,
		CreatedDefaultPath: xdgPath,
		CreatedExamplePath: examplePath,
	}, nil
}

func realExecutableDir() (string, error) {
	exePath, err := osExecutable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exePath)
	if err == nil {
		exePath = resolved
	}
	return filepath.Dir(exePath), nil
}

var (
	osExecutable    = os.Executable
	executableDir   = realExecutableDir
	osUserConfigDir = os.UserConfigDir
)

func EnsureConfigDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	return nil
}

func renderConfig(cfg Config) string {
	var b strings.Builder
	b.WriteString("# desktopify-lite config\n")
	b.WriteString("# Browser binary used when the prompt is left empty\n")
	b.WriteString("default_browser=")
	b.WriteString(cfg.DefaultBrowser)
	b.WriteString("\n")
	b.WriteString("# URL insertion template, must contain {url}\n")
	b.WriteString("default_url_template=")
	b.WriteString(cfg.DefaultURLTemplate)
	b.WriteString("\n")
	b.WriteString("# Extra browser flags applied when the prompt is left empty\n")
	b.WriteString("default_extra_flags=")
	b.WriteString(cfg.DefaultExtraFlags)
	b.WriteString("\n")
	b.WriteString("# Proxy for all network requests made by desktopify-lite (optional).\n")
	b.WriteString("# CLI override example: desktopify-lite --proxy=http://127.0.0.1:8080\n")
	b.WriteString("# Config value example: default_proxy=http://127.0.0.1:8080\n")
	b.WriteString("default_proxy=")
	b.WriteString(cfg.DefaultProxy)
	b.WriteString("\n")
	b.WriteString("# Disable Google favicon lookup during icon auto-discovery\n")
	if cfg.DisableGoogleFavicon {
		b.WriteString("disable_google_favicon=true\n")
	} else {
		b.WriteString("# disable_google_favicon=true\n")
	}
	if cfg.WithDebug {
		b.WriteString("with_debug=true\n")
	} else {
		b.WriteString("# with_debug=true\n")
	}
	return b.String()
}

func renderExampleConfig(cfg Config) string {
	var b strings.Builder
	b.WriteString("# Example desktopify-lite config\n")
	b.WriteString("# Copy this file to 'config' next to the binary for local development,\n")
	b.WriteString(fmt.Sprintf("# or to ~/.config/%s/%s/config for per-user settings.\n", orgConfigDirName, appConfigDirName))
	b.WriteString(renderConfig(cfg))
	return b.String()
}

func parseConfigFile(path string, base Config) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return base, fmt.Errorf("cannot open config: %w", err)
	}
	defer f.Close()

	cfg := base

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		key, val, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if !ok {
			switch key {
			case "with_debug":
				cfg.WithDebug = true
				continue
			default:
				continue
			}
		}

		val = trimQuotes(val)
		switch key {
		case "default_browser":
			if val != "" {
				cfg.DefaultBrowser = val
			}
		case "default_url_template":
			if val != "" {
				cfg.DefaultURLTemplate = val
			}
		case "default_extra_flags":
			cfg.DefaultExtraFlags = val
		case "default_proxy":
			cfg.DefaultProxy = val
		case "disable_google_favicon":
			parsed, err := parseBoolStrict(val)
			if err != nil {
				return base, fmt.Errorf("config: disable_google_favicon: %w", err)
			}
			cfg.DisableGoogleFavicon = parsed
		case "with_debug":
			parsed, err := parseBoolStrict(val)
			if err != nil {
				return base, fmt.Errorf("config: with_debug: %w", err)
			}
			cfg.WithDebug = parsed
		}
	}
	if err := s.Err(); err != nil {
		return base, fmt.Errorf("cannot read config: %w", err)
	}

	if strings.TrimSpace(cfg.DefaultBrowser) == "" {
		return base, errors.New("config: default_browser cannot be empty")
	}
	if strings.TrimSpace(cfg.DefaultURLTemplate) == "" {
		return base, errors.New("config: default_url_template cannot be empty")
	}
	if !strings.Contains(cfg.DefaultURLTemplate, "{url}") {
		return base, errors.New("config: default_url_template must contain {url}")
	}

	return cfg, nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func parseBoolStrict(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", s)
	}
}
