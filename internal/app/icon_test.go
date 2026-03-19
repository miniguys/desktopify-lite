package app

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildIconDownloadPlanFallbacks(t *testing.T) {
	plan, err := BuildIconDownloadPlan("https://sub.example.com/app", true)
	if err != nil {
		t.Fatal(err)
	}

	if len(plan) != 1 {
		t.Fatalf("len(plan)=%d, want 1", len(plan))
	}

	if plan[0].Origin != "google-host" {
		t.Fatalf("unexpected first origin: %q", plan[0].Origin)
	}
}

func TestBuildIconDownloadPlanCanDisableGoogleFallback(t *testing.T) {
	plan, err := BuildIconDownloadPlan("https://sub.example.com/app", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 0 {
		t.Fatalf("expected no plan when Google fallback is disabled, got %+v", plan)
	}
}

func TestBuildIconDownloadPlanSkipsGoogleForLocalHosts(t *testing.T) {
	for _, raw := range []string{
		"http://localhost:3000/app",
		"http://127.0.0.1:3000/app",
		"http://devbox.local/app",
		"http://router.home.arpa/app",
		"http://service.internal/app",
		"http://nas.lan/app",
	} {
		plan, err := BuildIconDownloadPlan(raw, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan) != 0 {
			t.Fatalf("expected no google fallback plan for %s, got %+v", raw, plan)
		}
	}
}

func TestBuildIconDownloadPlanDirectImage(t *testing.T) {
	plan, err := BuildIconDownloadPlan("https://example.com/assets/icon.svg", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 1 || plan[0].Ext != ".svg" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestExtractIconsFromHTML(t *testing.T) {
	baseURL, err := url.Parse("https://example.com/app/")
	if err != nil {
		t.Fatal(err)
	}

	doc := `
		<html><head>
			<link rel="icon" href="/favicon-32.png" sizes="32x32" type="image/png">
			<link rel="apple-touch-icon" href="icons/apple.svg" sizes="180x180">
		</head></html>`

	icons := extractIconsFromHTML(doc, baseURL)
	if len(icons) != 2 {
		t.Fatalf("len(icons)=%d, want 2", len(icons))
	}

	sortIconSources(icons)
	if icons[0].URL != "https://example.com/app/icons/apple.svg" {
		t.Fatalf("expected apple icon first, got %q", icons[0].URL)
	}
}

func TestExtractIconsFromManifest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manifest.json" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/manifest+json")
		_, _ = w.Write([]byte(`{
			"icons": [
				{"src":"/icons/icon-192.png","sizes":"192x192","type":"image/png","purpose":"any"},
				{"src":"/icons/icon-512.png","sizes":"512x512","type":"image/png","purpose":"maskable"},
				{"src":"","sizes":"64x64"}
			]
		}`))
	}))
	defer server.Close()

	icons, err := extractIconsFromManifest(server.URL+"/manifest.json", time.Second, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(icons) != 2 {
		t.Fatalf("len(icons)=%d, want 2", len(icons))
	}
	sortIconSources(icons)
	if icons[0].Purpose != "any" {
		t.Fatalf("expected purpose=any first, got %+v", icons)
	}
}

func TestSortIconSources(t *testing.T) {
	icons := []IconSource{
		{URL: "https://example.com/low.png", Ext: ".png", Origin: "html-link", Rel: "icon", Sizes: "32x32"},
		{URL: "https://example.com/high.svg", Ext: ".svg", Origin: "html-link", Rel: "apple-touch-icon", Sizes: "180x180"},
		{URL: "https://example.com/mid.png", Ext: ".png", Origin: "manifest", Purpose: "any", Sizes: "192x192"},
	}

	sortIconSources(icons)

	got := []string{icons[0].URL, icons[1].URL, icons[2].URL}
	want := []string{
		"https://example.com/high.svg",
		"https://example.com/mid.png",
		"https://example.com/low.png",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted order mismatch: got %v want %v", got, want)
		}
	}
}

func TestDetectIconExt(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		contentType string
		want        string
	}{
		{name: "ext from url", rawURL: "https://example.com/icon.webp", contentType: "image/png", want: ".webp"},
		{name: "ext from content type", rawURL: "https://example.com/icon", contentType: "image/svg+xml", want: ".svg"},
		{name: "jpeg content type", rawURL: "https://example.com/icon", contentType: "image/jpeg", want: ".jpg"},
		{name: "unknown", rawURL: "https://example.com/icon", contentType: "text/plain", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectIconExt(tc.rawURL, tc.contentType)
			if got != tc.want {
				t.Fatalf("detectIconExt(%q, %q)=%q, want %q", tc.rawURL, tc.contentType, got, tc.want)
			}
		})
	}
}

func TestNewHTTPClientProxy(t *testing.T) {
	client, err := newHTTPClient(time.Second, "http://127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type: %T", client.Transport)
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatal(err)
	}
	if proxyURL == nil || proxyURL.String() != "http://127.0.0.1:8080" {
		t.Fatalf("unexpected proxy URL: %v", proxyURL)
	}
}

func TestDownloadFileRejectsOversizedIcon(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte(strings.Repeat("a", maxIconBytes+1)))
	}))
	defer server.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "icon.png")

	_, err := downloadFile(server.URL+"/icon.png", dest, time.Second, "")
	if err == nil {
		t.Fatal("expected oversized icon error")
	}
	if !strings.Contains(err.Error(), "icon is too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadFileRejectsInvalidSVGContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write([]byte(`<html>not svg</html>`))
	}))
	defer server.Close()

	_, err := downloadFile(server.URL+"/icon.svg", filepath.Join(t.TempDir(), "icon.svg"), time.Second, "")
	if err == nil || !strings.Contains(err.Error(), "invalid svg content") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadFileAcceptsValidPNG(t *testing.T) {
	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR2WQAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngData)
	}))
	defer server.Close()

	path, err := downloadFile(server.URL+"/icon.png", filepath.Join(t.TempDir(), "icon.png"), time.Second, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, ".png") {
		t.Fatalf("unexpected saved path: %s", path)
	}
}

func TestDownloadFileAcceptsValidPNGWithOctetStreamContentType(t *testing.T) {
	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR2WQAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(pngData)
	}))
	defer server.Close()

	path, err := downloadFile(server.URL+"/icon", filepath.Join(t.TempDir(), "icon"), time.Second, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, ".png") {
		t.Fatalf("unexpected saved path: %s", path)
	}
}

func TestResolveIconExplicitRemoteURLWithoutExtensionUsesDirectDownload(t *testing.T) {
	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9WnR2WQAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/icon":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><link rel="icon" href="/fallback.png"></head></html>`))
		case "/fallback.png":
			t.Fatal("unexpected fallback request")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	baseDir := t.TempDir()
	paths := Paths{
		ApplicationsDir: filepath.Join(baseDir, "applications"),
		IconsDir:        filepath.Join(baseDir, "icons"),
	}
	if err := EnsureDirs(paths); err != nil {
		t.Fatal(err)
	}

	iconPath, err := ResolveIcon(Input{
		Name:            "Example App",
		IconURL:         server.URL + "/icon?id=123",
		IconURLExplicit: true,
	}, paths, DefaultConfig(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(iconPath, ".png") {
		t.Fatalf("expected .png icon path, got %q", iconPath)
	}
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(pngData) {
		t.Fatal("saved icon does not match direct download")
	}
}

func TestResolveIconCopiesLocalFile(t *testing.T) {
	baseDir := t.TempDir()
	paths := Paths{
		ApplicationsDir: filepath.Join(baseDir, "applications"),
		IconsDir:        filepath.Join(baseDir, "icons"),
	}
	if err := EnsureDirs(paths); err != nil {
		t.Fatal(err)
	}

	localIcon := filepath.Join(baseDir, "local-icon.svg")
	if err := os.WriteFile(localIcon, []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	iconPath, err := ResolveIcon(Input{Name: "Example App", IconURL: localIcon}, paths, DefaultConfig(), "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(iconPath, ".svg") {
		t.Fatalf("expected .svg icon path, got %q", iconPath)
	}
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<svg") {
		t.Fatal("expected copied svg content")
	}
}

func TestDownloadFileRejectsHTMLWithEmbeddedSVGWhenExpectingSVG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write([]byte(`<html><body><svg></svg></body></html>`))
	}))
	defer server.Close()

	_, err := downloadFile(server.URL+"/icon.svg", filepath.Join(t.TempDir(), "icon.svg"), time.Second, "")
	if err == nil || !strings.Contains(err.Error(), "invalid svg content") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveStaleIconVariantsRemovesAllWhenNoIconIsKept(t *testing.T) {
	baseDir := t.TempDir()
	paths := Paths{IconsDir: filepath.Join(baseDir, "icons")}
	if err := os.MkdirAll(paths.IconsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	for _, ext := range []string{".png", ".svg", ".ico"} {
		if err := os.WriteFile(filepath.Join(paths.IconsDir, "Example_App"+ext), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := removeStaleIconVariants(paths, "Example App", ""); err != nil {
		t.Fatal(err)
	}

	for _, ext := range []string{".png", ".svg", ".ico"} {
		if _, err := os.Stat(filepath.Join(paths.IconsDir, "Example_App"+ext)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, stat err=%v", ext, err)
		}
	}
}

func TestDefaultUserAgentUsesAppVersionWithoutPinnedChromeVersion(t *testing.T) {
	ua := defaultUserAgent()
	if !strings.Contains(ua, "desktopify-lite/"+version) {
		t.Fatalf("user agent should include app version, got %q", ua)
	}
	if strings.Contains(ua, "Chrome/") {
		t.Fatalf("user agent should not pin a Chrome version, got %q", ua)
	}
}

func TestResolveIconReportsFetchHTMLDocumentFailureWhenNoFallbacksExist(t *testing.T) {
	baseDir := t.TempDir()
	paths := Paths{
		ApplicationsDir: filepath.Join(baseDir, "applications"),
		IconsDir:        filepath.Join(baseDir, "icons"),
	}
	if err := EnsureDirs(paths); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveIcon(Input{Name: "Example App", IconURL: "http://127.0.0.1:1"}, paths, Config{DisableGoogleFavicon: true}, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "127.0.0.1:1") {
		t.Fatalf("expected error to mention failed source, got %v", err)
	}
	if strings.Contains(err.Error(), "no icon candidates found") {
		t.Fatalf("expected underlying fetch error, got %v", err)
	}
}
