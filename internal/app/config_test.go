package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config")
	content := strings.Join([]string{
		"default_browser=thorium",
		"default_url_template=--app={url}",
		"default_extra_flags=--profile-directory=Default --incognito",
		"default_proxy=http://127.0.0.1:8080",
		"with_debug",
	}, "\n") + "\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfigFile(p, DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultBrowser != "thorium" {
		t.Fatalf("DefaultBrowser=%q", cfg.DefaultBrowser)
	}
	if cfg.DefaultURLTemplate != "--app={url}" {
		t.Fatalf("DefaultURLTemplate=%q", cfg.DefaultURLTemplate)
	}
	if cfg.DefaultExtraFlags != "--profile-directory=Default --incognito" {
		t.Fatalf("DefaultExtraFlags=%q", cfg.DefaultExtraFlags)
	}
	if cfg.DefaultProxy != "http://127.0.0.1:8080" {
		t.Fatalf("DefaultProxy=%q", cfg.DefaultProxy)
	}
	if !cfg.WithDebug {
		t.Fatalf("WithDebug should be true")
	}
}

func TestRenderExampleConfigContainsHints(t *testing.T) {
	rendered := renderExampleConfig(DefaultConfig())
	for _, want := range []string{"Copy this file", "default_browser=", "default_url_template=", "default_extra_flags=", "default_proxy="} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered example config missing %q", want)
		}
	}
}
