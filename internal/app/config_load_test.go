package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigPrefersLocalConfigNearBinary(t *testing.T) {
	binDir := t.TempDir()
	userCfgDir := t.TempDir()
	localCfg := filepath.Join(binDir, localConfigFileName)
	if err := os.WriteFile(localCfg, []byte("default_browser=local-browser\ndefault_url_template=--app={url}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldExe := osExecutable
	oldUserCfg := osUserConfigDir
	defer func() {
		osExecutable = oldExe
		osUserConfigDir = oldUserCfg
	}()
	osExecutable = func() (string, error) { return filepath.Join(binDir, "desktopify-lite"), nil }
	osUserConfigDir = func() (string, error) { return userCfgDir, nil }

	cfg, meta, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultBrowser != "local-browser" {
		t.Fatalf("DefaultBrowser=%q", cfg.DefaultBrowser)
	}
	if cfg.DefaultURLTemplate != "--app={url}" {
		t.Fatalf("DefaultURLTemplate=%q", cfg.DefaultURLTemplate)
	}
	if meta.ActivePath != localCfg {
		t.Fatalf("ActivePath=%q, want %q", meta.ActivePath, localCfg)
	}
}

func TestLoadConfigCreatesDefaultAndExampleWhenMissing(t *testing.T) {
	binDir := t.TempDir()
	userCfgDir := t.TempDir()

	oldExe := osExecutable
	oldUserCfg := osUserConfigDir
	defer func() {
		osExecutable = oldExe
		osUserConfigDir = oldUserCfg
	}()
	osExecutable = func() (string, error) { return filepath.Join(binDir, "desktopify-lite"), nil }
	osUserConfigDir = func() (string, error) { return userCfgDir, nil }

	cfg, meta, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultBrowser == "" || cfg.DefaultURLTemplate == "" {
		t.Fatalf("unexpected empty config: %+v", cfg)
	}
	if meta.CreatedDefaultPath == "" || meta.CreatedExamplePath == "" {
		t.Fatalf("expected both default and example config to be created: %+v", meta)
	}
	if _, err := os.Stat(meta.CreatedDefaultPath); err != nil {
		t.Fatalf("default config missing: %v", err)
	}
	st, err := os.Stat(meta.CreatedExamplePath)
	if err != nil {
		t.Fatalf("example config missing: %v", err)
	}
	if st.Mode().Perm() != 0o444 {
		t.Fatalf("example config perms=%o, want 444", st.Mode().Perm())
	}
}

func TestLoadConfigRejectsInvalidBooleanValue(t *testing.T) {
	binDir := t.TempDir()
	userCfgDir := t.TempDir()
	localCfg := filepath.Join(binDir, localConfigFileName)
	if err := os.WriteFile(localCfg, []byte("default_browser=chromium\ndefault_url_template=--app={url}\nwith_debug=maybe\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldExe := osExecutable
	oldUserCfg := osUserConfigDir
	defer func() {
		osExecutable = oldExe
		osUserConfigDir = oldUserCfg
	}()
	osExecutable = func() (string, error) { return filepath.Join(binDir, "desktopify-lite"), nil }
	osUserConfigDir = func() (string, error) { return userCfgDir, nil }

	if _, _, err := LoadConfig(); err == nil {
		t.Fatal("expected invalid boolean value error")
	}
}
