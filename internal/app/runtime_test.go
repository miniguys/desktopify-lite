package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRuntimeOptionsProxy(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--proxy=http://127.0.0.1:8080"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandRun {
		t.Fatalf("Command=%q", opts.Command)
	}
	if opts.Proxy != "http://127.0.0.1:8080" {
		t.Fatalf("Proxy=%q", opts.Proxy)
	}
}

func TestParseRuntimeOptionsHelp(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--help"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandHelp {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsVersionFlag(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--version"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandVersion {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsVersionCommand(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"version"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandVersion {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsConfigUpdates(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"config", "--default_proxy=http://127.0.0.1:8080", "--with_debug=true", "--default_extra_flags="})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandConfig {
		t.Fatalf("Command=%q", opts.Command)
	}
	if opts.ConfigUpdates.DefaultProxy == nil || *opts.ConfigUpdates.DefaultProxy != "http://127.0.0.1:8080" {
		t.Fatalf("DefaultProxy=%v", opts.ConfigUpdates.DefaultProxy)
	}
	if opts.ConfigUpdates.WithDebug == nil || !*opts.ConfigUpdates.WithDebug {
		t.Fatalf("WithDebug=%v", opts.ConfigUpdates.WithDebug)
	}
	if opts.ConfigUpdates.DefaultExtraFlags == nil || *opts.ConfigUpdates.DefaultExtraFlags != "" {
		t.Fatalf("DefaultExtraFlags=%v", opts.ConfigUpdates.DefaultExtraFlags)
	}
}

func TestParseRuntimeOptionsConfigReset(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"config-reset"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandConfigReset {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsNonInteractive(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--url=https://example.com", "--name=Example", "--browser=chromium"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.NonInteractive {
		t.Fatal("expected non-interactive mode")
	}
	if opts.RunInput.URL != "https://example.com" || opts.RunInput.Name != "Example" || opts.RunInput.Browser != "chromium" {
		t.Fatalf("unexpected run input: %+v", opts.RunInput)
	}
}

func TestEffectiveProxyFlagOverridesConfig(t *testing.T) {
	got := EffectiveProxy(RuntimeOptions{Proxy: "http://127.0.0.1:8080"}, Config{DefaultProxy: "http://127.0.0.1:9090"})
	if got != "http://127.0.0.1:8080" {
		t.Fatalf("EffectiveProxy=%q", got)
	}
}

func TestValidateProxyURL(t *testing.T) {
	for _, tc := range []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "empty ok", raw: ""},
		{name: "http ok", raw: "http://127.0.0.1:8080"},
		{name: "https ok", raw: "https://proxy.local:8443"},
		{name: "missing scheme", raw: "127.0.0.1:8080", wantErr: true},
		{name: "unsupported scheme", raw: "socks5://127.0.0.1:1080", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProxyURL(tc.raw)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestDefaultPathsUsesXDGDataHome(t *testing.T) {
	old := os.Getenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		if old == "" {
			_ = os.Unsetenv("XDG_DATA_HOME")
		} else {
			_ = os.Setenv("XDG_DATA_HOME", old)
		}
	})

	xdg := t.TempDir()
	if err := os.Setenv("XDG_DATA_HOME", xdg); err != nil {
		t.Fatal(err)
	}

	paths, err := DefaultPaths()
	if err != nil {
		t.Fatal(err)
	}
	if paths.ApplicationsDir != filepath.Join(xdg, "applications") {
		t.Fatalf("ApplicationsDir=%q", paths.ApplicationsDir)
	}
	if paths.IconsDir != filepath.Join(xdg, "icons") {
		t.Fatalf("IconsDir=%q", paths.IconsDir)
	}
}

func TestParseRuntimeOptionsTracksExplicitIconURL(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--url=https://example.com", "--name=Example", "--icon-url=https://example.com/icon.png"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.RunInputExplicit.IconURL {
		t.Fatal("expected explicit icon-url to be tracked")
	}
}

func TestParseRuntimeOptionsRejectsInvalidBoolValue(t *testing.T) {
	if _, err := ParseRuntimeOptions([]string{"config", "--with_debug=maybe"}); err == nil {
		t.Fatal("expected error for invalid boolean value")
	}
}

func TestParseRuntimeOptionsSkipIcon(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"--url=https://example.com", "--name=Example", "--skip-icon"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.SkipIcon {
		t.Fatal("expected SkipIcon=true")
	}
}

func TestParseRuntimeOptionsRejectsSkipIconWithExplicitIconURL(t *testing.T) {
	if _, err := ParseRuntimeOptions([]string{"--url=https://example.com", "--name=Example", "--skip-icon", "--icon-url=https://example.com/icon.png"}); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestParseRuntimeOptionsVersionShortFlag(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"-v"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandVersion {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsConfigVersionFlag(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"config", "--version"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandVersion {
		t.Fatalf("Command=%q", opts.Command)
	}
}

func TestParseRuntimeOptionsConfigResetVersionFlag(t *testing.T) {
	opts, err := ParseRuntimeOptions([]string{"config-reset", "--version"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Command != CommandVersion {
		t.Fatalf("Command=%q", opts.Command)
	}
}
