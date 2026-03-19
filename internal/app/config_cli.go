package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConfigTargetPath(meta ConfigLoadMeta) (string, error) {
	if strings.TrimSpace(meta.ActivePath) != "" {
		return meta.ActivePath, nil
	}

	binDir, err := executableDir()
	if err != nil {
		return "", fmt.Errorf("resolve executable dir: %w", err)
	}
	localPath := filepath.Join(binDir, localConfigFileName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat local config: %w", err)
	}

	cfgDir, err := osUserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(cfgDir, orgConfigDirName, appConfigDirName, localConfigFileName), nil
}

func ApplyConfigUpdates(cfg Config, updates ConfigUpdates) (Config, error) {
	if updates.DefaultBrowser != nil {
		cfg.DefaultBrowser = *updates.DefaultBrowser
	}
	if updates.DefaultURLTemplate != nil {
		cfg.DefaultURLTemplate = *updates.DefaultURLTemplate
	}
	if updates.DefaultExtraFlags != nil {
		cfg.DefaultExtraFlags = *updates.DefaultExtraFlags
	}
	if updates.DefaultProxy != nil {
		cfg.DefaultProxy = *updates.DefaultProxy
	}
	if updates.DisableGoogleFavicon != nil {
		cfg.DisableGoogleFavicon = *updates.DisableGoogleFavicon
	}
	if updates.WithDebug != nil {
		cfg.WithDebug = *updates.WithDebug
	}

	if strings.TrimSpace(cfg.DefaultBrowser) == "" {
		return Config{}, fmt.Errorf("config: default_browser cannot be empty")
	}
	if strings.TrimSpace(cfg.DefaultURLTemplate) == "" {
		return Config{}, fmt.Errorf("config: default_url_template cannot be empty")
	}
	if !strings.Contains(cfg.DefaultURLTemplate, "{url}") {
		return Config{}, fmt.Errorf("config: default_url_template must contain {url}")
	}
	if err := ValidateProxyURL(cfg.DefaultProxy); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}

	return cfg, nil
}

func WriteConfigFile(path string, cfg Config) error {
	if err := EnsureConfigDir(filepath.Dir(path)); err != nil {
		return err
	}
	if err := osWriteFileAtomic(path, []byte(renderConfig(cfg)), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
