package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyConfigUpdates(t *testing.T) {
	browser := "thorium"
	urlTemplate := "--app={url}"
	extraFlags := "--incognito"
	proxy := "http://127.0.0.1:8080"
	withDebug := true

	cfg, err := ApplyConfigUpdates(DefaultConfig(), ConfigUpdates{
		DefaultBrowser:     &browser,
		DefaultURLTemplate: &urlTemplate,
		DefaultExtraFlags:  &extraFlags,
		DefaultProxy:       &proxy,
		WithDebug:          &withDebug,
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.DefaultBrowser != browser || cfg.DefaultURLTemplate != urlTemplate || cfg.DefaultExtraFlags != extraFlags || cfg.DefaultProxy != proxy || cfg.WithDebug != withDebug {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestRenderConfigContainsProxyExamples(t *testing.T) {
	rendered := renderConfig(DefaultConfig())
	for _, want := range []string{
		"desktopify-lite --proxy=http://127.0.0.1:8080",
		"default_proxy=http://127.0.0.1:8080",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered config missing %q", want)
		}
	}
}

func TestConfigTargetPathPrefersBrokenLocalConfigWithoutParsing(t *testing.T) {
	binDir := t.TempDir()
	userCfgDir := t.TempDir()
	localPath := filepath.Join(binDir, localConfigFileName)
	if err := os.WriteFile(localPath, []byte("default_url_template=broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldExecDir := executableDir
	oldUserCfg := osUserConfigDir
	t.Cleanup(func() {
		executableDir = oldExecDir
		osUserConfigDir = oldUserCfg
	})

	executableDir = func() (string, error) { return binDir, nil }
	osUserConfigDir = func() (string, error) { return userCfgDir, nil }

	targetPath, err := ConfigTargetPath(ConfigLoadMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if targetPath != localPath {
		t.Fatalf("targetPath=%q, want %q", targetPath, localPath)
	}
}

func TestRunConfigCommandRepairsBrokenConfig(t *testing.T) {
	binDir := t.TempDir()
	userCfgDir := t.TempDir()
	localPath := filepath.Join(binDir, localConfigFileName)
	if err := os.WriteFile(localPath, []byte("with_debug=definitely\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldExecDir := executableDir
	oldUserCfg := osUserConfigDir
	t.Cleanup(func() {
		executableDir = oldExecDir
		osUserConfigDir = oldUserCfg
	})

	executableDir = func() (string, error) { return binDir, nil }
	osUserConfigDir = func() (string, error) { return userCfgDir, nil }

	browser := "firefox"
	if err := runConfigCommand(RuntimeOptions{ConfigUpdates: ConfigUpdates{DefaultBrowser: &browser}}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "default_browser=firefox") {
		t.Fatalf("updated config missing browser override:\n%s", text)
	}
	if !strings.Contains(text, "default_url_template=--app={url}") {
		t.Fatalf("updated config should be rebuilt from defaults when source is invalid:\n%s", text)
	}
}
