package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveIconRefreshesExistingIconWithSameExtension(t *testing.T) {
	baseDir := t.TempDir()
	paths := Paths{
		ApplicationsDir: filepath.Join(baseDir, "applications"),
		IconsDir:        filepath.Join(baseDir, "icons"),
	}
	if err := EnsureDirs(paths); err != nil {
		t.Fatal(err)
	}

	responses := map[string][]byte{
		"/icon.png": solidPNG(t, color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(responses[r.URL.Path])
	}))
	defer server.Close()

	cfg := DefaultConfig()
	in := Input{Name: "Example App", IconURL: server.URL + "/icon.png"}

	iconPath, err := ResolveIcon(in, paths, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, responses["/icon.png"]) {
		t.Fatal("first icon contents were not written as served")
	}

	responses["/icon.png"] = solidPNG(t, color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})
	iconPath, err = ResolveIcon(in, paths, cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, responses["/icon.png"]) {
		t.Fatal("refreshed icon contents were not updated")
	}
}

func solidPNG(t *testing.T, c color.RGBA) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestBuildAndWriteDesktopEntryEndToEnd(t *testing.T) {
	baseDir := t.TempDir()
	paths := Paths{
		ApplicationsDir: filepath.Join(baseDir, "applications"),
		IconsDir:        filepath.Join(baseDir, "icons"),
	}
	if err := EnsureDirs(paths); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	iconPath := filepath.Join(paths.IconsDir, "Example_App.png")
	in := Input{
		URL:            "https://example.com/app?a=1&b=2",
		Name:           "Example App",
		Browser:        "chromium",
		URLTemplate:    "--app={url}",
		ExtraFlags:     "--profile-directory=Default",
		StartupWMClass: "ExampleApp",
	}

	entry, filename, err := BuildDesktopEntry(in, iconPath, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if filename != "Example_App.desktop" {
		t.Fatalf("filename=%q", filename)
	}

	desktopPath := paths.DesktopFilePath(filename)
	if err := WriteDesktopFile(desktopPath, entry); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	checks := []string{
		"[Desktop Entry]",
		"Name=Example App",
		"Exec=chromium --profile-directory=Default \"--app=https://example.com/app?a=1&b=2\"",
		"StartupWMClass=ExampleApp",
		"Icon=" + iconPath,
	}
	for _, want := range checks {
		if !strings.Contains(text, want) {
			t.Fatalf("desktop file missing %q in:\n%s", want, text)
		}
	}
}
