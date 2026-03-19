package app

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAskInputIncludesStartupWMClassAndProxy(t *testing.T) {
	cfg := Config{
		DefaultBrowser:     "chromium",
		DefaultURLTemplate: "--app={url}",
	}

	reader := bufio.NewReader(strings.NewReader(strings.Join([]string{
		"example.com",
		"",
		"Example App",
		"",
		"",
		"--incognito",
		"ExampleApp",
		"http://127.0.0.1:8080",
	}, "\n") + "\n"))

	in, err := AskInput(reader, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if in.URL != "https://example.com" {
		t.Fatalf("URL=%q", in.URL)
	}
	if in.IconURL != in.URL {
		t.Fatalf("IconURL=%q, want website URL", in.IconURL)
	}
	if in.IconURLExplicit {
		t.Fatal("expected auto icon URL to be implicit")
	}
	if in.StartupWMClass != "ExampleApp" {
		t.Fatalf("StartupWMClass=%q", in.StartupWMClass)
	}
	if in.Proxy != "http://127.0.0.1:8080" {
		t.Fatalf("Proxy=%q", in.Proxy)
	}
}

func TestNormalizeAndValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "adds https", input: "example.com", want: "https://example.com"},
		{name: "adds https to host port", input: "example.com:8080", want: "https://example.com:8080"},
		{name: "defaults localhost port to http", input: "localhost:3000", want: "http://localhost:3000"},
		{name: "defaults ip port to http", input: "127.0.0.1:8080", want: "http://127.0.0.1:8080"},
		{name: "keeps scheme", input: "https://example.com/app", want: "https://example.com/app"},
		{name: "rejects empty", input: "", wantErr: true},
		{name: "rejects malformed", input: "https://", wantErr: true},
		{name: "rejects unsupported scheme", input: "ftp://example.com", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeAndValidateURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeAndValidateIconLocation(t *testing.T) {
	localFile := filepath.Join(t.TempDir(), "icon.svg")
	if err := os.WriteFile(localFile, []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "https url", input: "https://example.com/icon.png", want: "https://example.com/icon.png"},
		{name: "file url", input: "file:///tmp/icon.svg", want: filepath.Clean("/tmp/icon.svg")},
		{name: "existing local file", input: localFile, want: filepath.Clean(localFile)},
		{name: "rejects bad file url", input: "file://remote-host/icon.svg", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeAndValidateIconLocation(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
