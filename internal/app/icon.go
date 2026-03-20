package app

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var errInvalidIconURL = errors.New("invalid icon url")

const maxIconBytes = 10 << 20

type IconSource struct {
	URL     string
	Ext     string
	Origin  string
	Rel     string
	Sizes   string
	Purpose string
}

type manifestFile struct {
	Icons []manifestIcon `json:"icons"`
}

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose"`
}

func ResolveIcon(in Input, paths Paths, cfg Config, proxyURL string) (string, error) {
	iconInput := strings.TrimSpace(in.IconURL)
	if iconInput == "" {
		return "", errors.New("icon url cannot be empty")
	}

	if cfg.WithDebug {
		fmt.Println(styleInfoName.Render("                         ━━━━━━       "))
		fmt.Println(styleProcess.Render("                     Loading icon...  "))
	}

	stem := launcherFileStem(in.Name)
	destStem := filepath.Join(paths.IconsDir, stem)

	if localPath, ok, err := resolveLocalIconSource(iconInput); ok || err != nil {
		if err != nil {
			return "", err
		}
		return copyLocalIcon(localPath, destStem, 15*time.Second)
	}

	parsed, err := url.Parse(iconInput)
	if err != nil || parsed.Host == "" {
		return "", errInvalidIconURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errInvalidIconURL
	}

	if in.IconURLExplicit {
		return downloadFile(iconInput, destStem, 15*time.Second, proxyURL)
	}

	var errs []string

	tryDownload := func(candidate IconSource) (string, error) {
		localIconPath := destStem
		if candidate.Ext != "" {
			localIconPath += candidate.Ext
		}
		return downloadFile(candidate.URL, localIconPath, 15*time.Second, proxyURL)
	}

	htmlDoc, baseURL, err := fetchHTMLDocument(iconInput, 15*time.Second, proxyURL)
	if err != nil {
		errs = append(errs, fmt.Sprintf("%s: %v", iconInput, err))
	} else {
		htmlIcons := extractIconsFromHTML(htmlDoc, baseURL)
		sortIconSources(htmlIcons)
		for _, src := range htmlIcons {
			saved, err := tryDownload(src)
			if err == nil {
				return saved, nil
			}
			errs = append(errs, fmt.Sprintf("%s: %v", src.URL, err))
		}

		manifestURL := extractManifestURL(htmlDoc, baseURL)
		if manifestURL != "" {
			manifestIcons, err := extractIconsFromManifest(manifestURL, 15*time.Second, proxyURL)
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", manifestURL, err))
			} else {
				sortIconSources(manifestIcons)
				for _, src := range manifestIcons {
					saved, err := tryDownload(src)
					if err == nil {
						return saved, nil
					}
					errs = append(errs, fmt.Sprintf("%s: %v", src.URL, err))
				}
			}
		}
	}

	plan, err := BuildIconDownloadPlan(iconInput, !cfg.DisableGoogleFavicon)
	if err != nil {
		return "", err
	}
	for _, candidate := range plan {
		saved, err := tryDownload(candidate)
		if err == nil {
			return saved, nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", candidate.URL, err))
	}

	if len(errs) == 0 {
		return "", errors.New("download icon: no icon candidates found")
	}

	return "", fmt.Errorf("download icon: all icon sources failed: %s", strings.Join(errs, " | "))
}

func BuildIconDownloadPlan(iconURL string, allowGoogleFallback bool) ([]IconSource, error) {
	parsed, err := url.Parse(iconURL)
	if err != nil || parsed.Host == "" {
		return nil, errInvalidIconURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errInvalidIconURL
	}

	if ext := normalizeIconExt(filepath.Ext(parsed.Path)); ext != "" {
		return []IconSource{{
			URL:    iconURL,
			Ext:    ext,
			Origin: "direct",
		}}, nil
	}

	host := parsed.Hostname()
	if !allowGoogleFallback || !shouldUseGoogleFaviconFallback(host) {
		return nil, nil
	}

	return []IconSource{{
		URL:    "https://www.google.com/s2/favicons?domain=" + host + "&sz=256",
		Ext:    ".png",
		Origin: "google-host",
	}}, nil
}

func resolveLocalIconSource(raw string) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, nil
	}

	if strings.HasPrefix(strings.ToLower(raw), "file://") {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "file" {
			return "", false, errInvalidIconURL
		}
		if parsed.Host != "" && parsed.Host != "localhost" {
			return "", false, errInvalidIconURL
		}
		path, err := url.PathUnescape(parsed.Path)
		if err != nil || strings.TrimSpace(path) == "" {
			return "", false, errInvalidIconURL
		}
		return filepath.Clean(path), true, nil
	}

	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme != "" {
		if parsed.Scheme == "http" || parsed.Scheme == "https" {
			return "", false, nil
		}
		return "", false, errInvalidIconURL
	}

	if strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false, fmt.Errorf("resolve icon path: %w", err)
		}
		return filepath.Join(home, raw[2:]), true, nil
	}

	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../") {
		return filepath.Clean(raw), true, nil
	}

	if info, err := os.Stat(raw); err == nil && !info.IsDir() {
		return filepath.Clean(raw), true, nil
	}

	if normalizeIconExt(filepath.Ext(raw)) != "" && !strings.Contains(raw, "/") && !strings.Contains(raw, `\`) {
		return filepath.Clean(raw), true, nil
	}

	return "", false, nil
}

func copyLocalIcon(sourcePath, destStem string, timeout time.Duration) (string, error) {
	_ = timeout
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return "", fmt.Errorf("load local icon: %w", err)
	}
	if len(data) > maxIconBytes {
		return "", fmt.Errorf("save icon: icon is too large")
	}

	return saveIconBytes(data, sourcePath, destStem, "")
}

func fetchHTMLDocument(rawURL string, timeout time.Duration, proxyURL string) (string, *url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", nil, errInvalidIconURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", nil, errInvalidIconURL
	}

	client, err := newHTTPClient(timeout, proxyURL)
	if err != nil {
		return "", nil, fmt.Errorf("download icon: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("fetch html: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("fetch html: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("fetch html: http %d", resp.StatusCode)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" &&
		!strings.Contains(contentType, "text/html") &&
		!strings.Contains(contentType, "application/xhtml+xml") {
		return "", nil, fmt.Errorf("fetch html: unexpected content type %q", resp.Header.Get("Content-Type"))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", nil, fmt.Errorf("fetch html: %w", err)
	}

	base := resp.Request.URL
	if base == nil {
		base = parsed
	}

	return string(data), base, nil
}

func extractIconsFromHTML(doc string, baseURL *url.URL) []IconSource {
	tagRe := regexp.MustCompile(`(?is)<link\b[^>]*>`)
	attrRe := regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)

	matches := tagRe.FindAllString(doc, -1)
	out := make([]IconSource, 0, len(matches))

	for _, tag := range matches {
		attrs := map[string]string{}
		for _, m := range attrRe.FindAllStringSubmatch(tag, -1) {
			key := strings.ToLower(strings.TrimSpace(m[1]))
			val := ""
			switch {
			case m[3] != "":
				val = m[3]
			case m[4] != "":
				val = m[4]
			default:
				val = m[5]
			}
			attrs[key] = strings.TrimSpace(val)
		}

		rel := strings.ToLower(attrs["rel"])
		href := attrs["href"]
		if href == "" || !isIconRel(rel) {
			continue
		}

		resolved := resolveURL(baseURL, href)
		if resolved == "" {
			continue
		}

		ext := detectIconExt(resolved, attrs["type"])
		if ext == "" {
			ext = ".png"
		}

		out = append(out, IconSource{
			URL:    resolved,
			Ext:    ext,
			Origin: "html-link",
			Rel:    rel,
			Sizes:  attrs["sizes"],
		})
	}

	return dedupeIconSources(out)
}

func extractManifestURL(doc string, baseURL *url.URL) string {
	tagRe := regexp.MustCompile(`(?is)<link\b[^>]*>`)
	attrRe := regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)

	for _, tag := range tagRe.FindAllString(doc, -1) {
		attrs := map[string]string{}
		for _, m := range attrRe.FindAllStringSubmatch(tag, -1) {
			key := strings.ToLower(strings.TrimSpace(m[1]))
			val := ""
			switch {
			case m[3] != "":
				val = m[3]
			case m[4] != "":
				val = m[4]
			default:
				val = m[5]
			}
			attrs[key] = strings.TrimSpace(val)
		}

		if strings.ToLower(attrs["rel"]) != "manifest" || attrs["href"] == "" {
			continue
		}

		return resolveURL(baseURL, attrs["href"])
	}

	return ""
}

func extractIconsFromManifest(manifestURL string, timeout time.Duration, proxyURL string) ([]IconSource, error) {
	client, err := newHTTPClient(timeout, proxyURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Accept", "application/manifest+json,application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch manifest: http %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}

	var mf manifestFile
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}

	baseURL, err := url.Parse(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}

	out := make([]IconSource, 0, len(mf.Icons))
	for _, icon := range mf.Icons {
		if strings.TrimSpace(icon.Src) == "" {
			continue
		}

		resolved := resolveURL(baseURL, icon.Src)
		if resolved == "" {
			continue
		}

		ext := detectIconExt(resolved, icon.Type)
		if ext == "" {
			ext = ".png"
		}

		out = append(out, IconSource{
			URL:     resolved,
			Ext:     ext,
			Origin:  "manifest",
			Sizes:   icon.Sizes,
			Purpose: icon.Purpose,
		})
	}

	return dedupeIconSources(out), nil
}

func isIconRel(rel string) bool {
	rel = strings.ToLower(strings.TrimSpace(rel))
	if rel == "" {
		return false
	}

	parts := strings.Fields(rel)
	for _, part := range parts {
		switch part {
		case "icon", "shortcut", "apple-touch-icon", "apple-touch-icon-precomposed", "mask-icon":
			return true
		}
	}

	return false
}

func iconTypePriority(src IconSource) int {
	rel := strings.ToLower(src.Rel)
	purpose := strings.ToLower(strings.TrimSpace(src.Purpose))
	switch {
	case src.Origin == "direct":
		return 0
	case strings.Contains(rel, "apple-touch-icon"):
		return 1
	case strings.Contains(purpose, "any"):
		return 2
	case strings.Contains(purpose, "maskable"):
		return 3
	case strings.Contains(rel, "icon"):
		return 4
	case strings.Contains(rel, "mask-icon"):
		return 5
	default:
		return 6
	}
}

func iconSizePriority(s string) int {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "any" {
		return 1 << 30
	}
	best := 0
	for _, part := range strings.Fields(s) {
		dims := strings.Split(part, "x")
		if len(dims) != 2 {
			continue
		}
		w, err1 := strconv.Atoi(dims[0])
		h, err2 := strconv.Atoi(dims[1])
		if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
			continue
		}
		if w < h {
			if w > best {
				best = w
			}
		} else if h > best {
			best = h
		}
	}
	return best
}

func iconExtPriority(ext string) int {
	switch strings.ToLower(ext) {
	case ".svg":
		return 0
	case ".png":
		return 1
	case ".jpg", ".jpeg":
		return 2
	case ".ico":
		return 3
	case ".webp":
		return 4
	default:
		return 5
	}
}

func compareIconSources(a, b IconSource) bool {
	if ta, tb := iconTypePriority(a), iconTypePriority(b); ta != tb {
		return ta < tb
	}
	if sa, sb := iconSizePriority(a.Sizes), iconSizePriority(b.Sizes); sa != sb {
		return sa > sb
	}
	if ea, eb := iconExtPriority(a.Ext), iconExtPriority(b.Ext); ea != eb {
		return ea < eb
	}
	return a.URL < b.URL
}

func detectIconExt(rawURL, contentType string) string {
	if ext := normalizeIconExt(filepath.Ext(rawURL)); ext != "" {
		return ext
	}

	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(contentType, "image/svg"):
		return ".svg"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/x-icon"),
		strings.Contains(contentType, "image/vnd.microsoft.icon"):
		return ".ico"
	default:
		return ""
	}
}

func normalizeIconExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".svg", ".png", ".jpg", ".jpeg", ".ico":
		return ext
	default:
		return ""
	}
}

func resolveURL(baseURL *url.URL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	resolved := baseURL.ResolveReference(parsed)
	if resolved == nil || (resolved.Scheme != "http" && resolved.Scheme != "https") || resolved.Host == "" {
		return ""
	}

	return resolved.String()
}

func sortIconSources(items []IconSource) {
	sort.SliceStable(items, func(i, j int) bool {
		return compareIconSources(items[i], items[j])
	})
}

func dedupeIconSources(items []IconSource) []IconSource {
	seen := make(map[string]struct{}, len(items))
	out := make([]IconSource, 0, len(items))
	for _, item := range items {
		key := item.URL + "\x00" + item.Ext + "\x00" + item.Rel + "\x00" + item.Sizes + "\x00" + item.Purpose
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func downloadFile(rawURL, destPath string, timeout time.Duration, proxyURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", errInvalidIconURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errInvalidIconURL
	}

	client, err := newHTTPClient(timeout, proxyURL)
	if err != nil {
		return "", fmt.Errorf("download icon: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("download icon: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent())
	req.Header.Set("Accept", "image/*,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download icon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download icon: http %d", resp.StatusCode)
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxIconBytes+1))
	if err != nil {
		return "", fmt.Errorf("save icon: %w", err)
	}
	if len(data) > maxIconBytes {
		return "", fmt.Errorf("save icon: icon is too large")
	}

	return saveIconBytes(data, rawURL, destPath, contentType)
}

func validateIconContent(data []byte, ext, contentType string) error {
	if len(data) == 0 {
		return errors.New("icon is empty")
	}

	switch normalizeIconExt(ext) {
	case ".svg":
		dec := xml.NewDecoder(bytes.NewReader(data))
		for {
			tok, err := dec.Token()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return errors.New("invalid svg content")
				}
				return errors.New("invalid svg content")
			}
			if start, ok := tok.(xml.StartElement); ok {
				if strings.EqualFold(start.Name.Local, "svg") {
					return nil
				}
				return errors.New("invalid svg content")
			}
		}
	case ".ico":
		if len(data) < 4 || !(data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00) {
			return errors.New("invalid ico content")
		}
		return nil
	case ".webp":
		if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
			return errors.New("invalid webp content")
		}
		return nil
	case ".png", ".jpg", ".jpeg":
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("invalid raster image: %w", err)
		}
		if cfg.Width <= 0 || cfg.Height <= 0 {
			return errors.New("invalid raster image dimensions")
		}
		if normalizeIconExt(ext) == ".png" && format != "png" {
			return fmt.Errorf("content does not match %s", ext)
		}
		if (normalizeIconExt(ext) == ".jpg" || normalizeIconExt(ext) == ".jpeg") && format != "jpeg" {
			return fmt.Errorf("content does not match %s", ext)
		}
		return nil
	default:
		if strings.Contains(contentType, "image/") {
			return nil
		}
		return errors.New("unsupported icon format")
	}
}

func saveIconBytes(data []byte, sourceRef, destPath, contentType string) (string, error) {
	finalExt := normalizeIconExt(filepath.Ext(destPath))
	if finalExt == "" {
		finalExt = detectIconExt(sourceRef, contentType)
		if finalExt == "" {
			finalExt = detectIconExtFromData(data)
		}
		if finalExt == "" {
			finalExt = ".png"
		}
		destPath = destPath + finalExt
	}

	if err := validateIconContent(data, finalExt, contentType); err != nil {
		return "", fmt.Errorf("save icon: %w", err)
	}

	if err := osMkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return "", err
	}

	if err := osWriteFileAtomic(destPath, data, 0o644); err != nil {
		return "", fmt.Errorf("save icon: %w", err)
	}

	return destPath, nil
}

func detectIconExtFromData(data []byte) string {
	if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00 {
		return ".ico"
	}
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return ".webp"
	}
	if len(data) >= 4 && string(data[:4]) == "\x89PNG" {
		return ".png"
	}
	if len(data) >= 2 && data[0] == 0xff && data[1] == 0xd8 {
		return ".jpg"
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 {
		dec := xml.NewDecoder(bytes.NewReader(trimmed))
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			if start, ok := tok.(xml.StartElement); ok {
				if strings.EqualFold(start.Name.Local, "svg") {
					return ".svg"
				}
				break
			}
		}
	}
	return ""
}

func newHTTPClient(timeout time.Duration, proxyURL string) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

func isLocalHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if host == "localhost" {
		return true
	}
	if net.ParseIP(host) != nil {
		return true
	}
	for _, suffix := range []string{".local", ".lan", ".internal", ".home.arpa", ".localdomain"} {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

func shouldUseGoogleFaviconFallback(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || isLocalHost(host) {
		return false
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if !strings.Contains(host, ".") {
		return false
	}

	labels := strings.Split(host, ".")
	tld := labels[len(labels)-1]
	if tld == "" {
		return false
	}
	if tld == "test" || tld == "example" || tld == "invalid" || tld == "localhost" || tld == "local" || tld == "internal" || tld == "lan" || tld == "arpa" {
		return false
	}
	if strings.HasPrefix(tld, "xn--") {
		return true
	}
	for _, r := range tld {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return len(tld) >= 2
}

func removeStaleIconVariants(paths Paths, appName string, keepPath string) error {
	stem := launcherFileStem(appName)
	keepBase := filepath.Base(keepPath)
	for _, ext := range []string{".svg", ".png", ".jpg", ".jpeg", ".webp", ".ico"} {
		candidate := filepath.Join(paths.IconsDir, stem+ext)
		if filepath.Base(candidate) == keepBase {
			continue
		}
		if err := os.Remove(candidate); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func defaultUserAgent() string {
	return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Safari/537.36 desktopify-lite/" + version
}

var osMkdirAll = func(p string, perm os.FileMode) error { return os.MkdirAll(p, perm) }
