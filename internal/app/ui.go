package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/miniguys/desktopify-lite/internal/lipgloss"
)

func AskInput(r *bufio.Reader, cfg Config) (Input, error) {
	fmt.Println(styleInfoName.Render("                         ━━━━━━       "))
	fmt.Println()

	websiteURL, err := askLine(r, questionURL())
	if err != nil {
		return Input{}, err
	}
	websiteURL, err = normalizeAndValidateURL(websiteURL)
	if err != nil {
		return Input{}, err
	}

	iconURLRaw, err := askLine(r, questionIconURL())
	if err != nil {
		return Input{}, err
	}
	iconURLExplicit := strings.TrimSpace(iconURLRaw) != ""
	iconURL := websiteURL
	if iconURLExplicit {
		iconURL, err = normalizeAndValidateIconLocation(iconURLRaw)
		if err != nil {
			return Input{}, fmt.Errorf("invalid icon url: %w", err)
		}
	}

	name, err := askLine(r, questionName())
	if err != nil {
		return Input{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Input{}, errors.New("name cannot be empty")
	}

	browser, err := askLine(r, styledQuestion("Browser binary", "default: "+cfg.DefaultBrowser))
	if err != nil {
		return Input{}, err
	}

	urlTemplate, err := askLine(r, styledQuestion("URL flag template", "default: "+cfg.DefaultURLTemplate))
	if err != nil {
		return Input{}, err
	}

	extraHint := "optional"
	if cfg.DefaultExtraFlags != "" {
		extraHint = "default: " + cfg.DefaultExtraFlags
	}
	extraFlags, err := askLine(r, styledQuestion("Extra flags", extraHint))
	if err != nil {
		return Input{}, err
	}

	startupWMClass, err := askLine(r, styledQuestion("StartupWMClass", "optional"))
	if err != nil {
		return Input{}, err
	}

	proxyHint := "optional"
	if cfg.DefaultProxy != "" {
		proxyHint = "default: " + cfg.DefaultProxy
	}
	proxy, err := askLine(r, styledQuestion("Proxy URL", proxyHint))
	if err != nil {
		return Input{}, err
	}
	if err := ValidateProxyURL(proxy); err != nil {
		return Input{}, err
	}

	return Input{
		URL:             websiteURL,
		IconURL:         iconURL,
		Name:            name,
		Browser:         browser,
		URLTemplate:     urlTemplate,
		ExtraFlags:      extraFlags,
		StartupWMClass:  startupWMClass,
		Proxy:           proxy,
		IconURLExplicit: iconURLExplicit,
	}, nil
}

func askLine(r *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)

	s, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			if len(s) == 0 {
				return "", errCanceled
			}
			return strings.TrimSpace(s), nil
		}
		return "", err
	}

	return strings.TrimSpace(s), nil
}

func questionURL() string {
	pref := styleInfoValue.Render(" >> ")
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		pref,
		styleInputs.Render("Website URL ")+styleHints.Render("(e.g. google.com)"),
	) + styleInputs.Render(": ")
}

func questionIconURL() string {
	pref := styleInfoValue.Render(" >> ")
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		pref,
		styleInputs.Render("Icon URL ")+styleHints.Render("(Leave empty to auto-fetch)"),
	) + styleInputs.Render(": ")
}

func questionName() string {
	pref := styleInfoValue.Render(" >> ")
	return lipgloss.JoinHorizontal(lipgloss.Left, pref, styleInputs.Render("WebApp name")) + styleInputs.Render(": ")
}

func normalizeAndValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url cannot be empty")
	}

	if !strings.Contains(raw, "://") {
		scheme := "https://"
		if shouldDefaultToHTTP(raw) {
			scheme = "http://"
		}
		raw = scheme + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("malformed url")
	}

	switch parsed.Scheme {
	case "http", "https":
	default:
		return "", errors.New("url must use http or https")
	}

	if parsed.Hostname() == "" {
		return "", errors.New("url must include a valid host")
	}

	return parsed.String(), nil
}

func shouldDefaultToHTTP(raw string) bool {
	parsed, err := url.Parse("//" + raw)
	if err != nil {
		return false
	}

	host := parsed.Hostname()
	if host == "" {
		return false
	}

	if strings.EqualFold(host, "localhost") {
		return true
	}

	return net.ParseIP(host) != nil
}

func normalizeAndValidateIconLocation(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url cannot be empty")
	}

	if path, ok, err := normalizeLocalIconPath(raw); ok || err != nil {
		return path, err
	}

	return normalizeAndValidateURL(raw)
}

func normalizeLocalIconPath(raw string) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, nil
	}

	if strings.HasPrefix(strings.ToLower(raw), "file://") {
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme != "file" {
			return "", false, errors.New("malformed file url")
		}
		if parsed.Host != "" && parsed.Host != "localhost" {
			return "", false, errors.New("file url must not contain a remote host")
		}
		path, err := url.PathUnescape(parsed.Path)
		if err != nil || strings.TrimSpace(path) == "" {
			return "", false, errors.New("malformed file url")
		}
		return filepath.Clean(path), true, nil
	}

	if strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false, err
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

func styledQuestion(text, hint string) string {
	pref := styleInfoValue.Render(" >> ")

	questionPart := styleInputs.Render(text)
	hintPart := ""
	if hint != "" {
		hintPart = styleHints.Render(fmt.Sprintf(" (%s)", hint))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		pref,
		questionPart,
		hintPart,
	) + styleInputs.Render(": ")
}
