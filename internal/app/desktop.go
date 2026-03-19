package app

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"unicode"
)

func BuildDesktopEntry(in Input, iconRef string, cfg Config) (desktopEntry string, filename string, _ error) {
	if strings.TrimSpace(in.Name) == "" {
		return "", "", errors.New("name cannot be empty")
	}
	if strings.TrimSpace(in.URL) == "" {
		return "", "", errors.New("url cannot be empty")
	}

	fileStem := launcherFileStem(in.Name)
	filename = fileStem + ".desktop"

	exec, err := BuildExec(in, cfg)
	if err != nil {
		return "", "", err
	}

	browser := strings.TrimSpace(in.Browser)
	if browser == "" {
		browser = strings.TrimSpace(cfg.DefaultBrowser)
	}

	lines := []string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=" + escapeDesktopValue(in.Name),
	}
	if browser != "" {
		lines = append(lines, "TryExec="+escapeDesktopValue(browser))
	}
	lines = append(lines,
		"Exec="+exec,
	)
	if strings.TrimSpace(iconRef) != "" {
		lines = append(lines, "Icon="+escapeDesktopValue(iconRef))
	}
	if strings.TrimSpace(in.StartupWMClass) != "" {
		lines = append(lines, "StartupWMClass="+escapeDesktopValue(in.StartupWMClass))
	}
	lines = append(lines,
		"Terminal=false",
		"Categories=Network;WebBrowser;",
		"",
	)

	desktopEntry = strings.Join(lines, "\n")
	return desktopEntry, filename, nil
}

func WriteDesktopFile(path string, content string) error {
	if err := osWriteFileAtomic(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("cannot write desktop file: %w", err)
	}
	return nil
}

func launcherFileStem(s string) string {
	base := sanitizeFileStem(s)
	if !needsLauncherStemDisambiguation(s, base) {
		return base
	}
	return fmt.Sprintf("%s-%08x", base, launcherStemHash(s))
}

func needsLauncherStemDisambiguation(original, sanitized string) bool {
	trimmed := strings.TrimSpace(original)
	if trimmed == "" {
		return false
	}
	if sanitized == "webapp" {
		return true
	}
	for _, r := range trimmed {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), unicode.IsSpace(r), r == '-', r == '_':
			continue
		default:
			return true
		}
	}
	return false
}

func launcherStemHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func sanitizeFileStem(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r),
			unicode.IsDigit(r),
			r == '-', r == '_':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "webapp"
	}
	return out
}

func BuildExec(in Input, cfg Config) (string, error) {
	argv, err := BuildExecArgs(in, cfg)
	if err != nil {
		return "", err
	}
	return joinExecArgs(argv), nil
}

func BuildExecArgs(in Input, cfg Config) ([]string, error) {
	browser := strings.TrimSpace(in.Browser)
	if browser == "" {
		browser = cfg.DefaultBrowser
	}
	if browser == "" {
		return nil, errors.New("browser binary cannot be empty")
	}

	template := strings.TrimSpace(in.URLTemplate)
	if template == "" {
		template = cfg.DefaultURLTemplate
	}
	if !strings.Contains(template, "{url}") {
		return nil, errors.New("url template must contain {url}")
	}

	extra := strings.TrimSpace(in.ExtraFlags)
	if extra == "" {
		extra = cfg.DefaultExtraFlags
	}

	args := []string{browser}

	if extra != "" {
		extraArgs, err := splitCommandLine(extra)
		if err != nil {
			return nil, fmt.Errorf("invalid extra flags: %w", err)
		}
		args = append(args, extraArgs...)
	}

	templateArgs, err := splitCommandLine(strings.ReplaceAll(template, "{url}", in.URL))
	if err != nil {
		return nil, fmt.Errorf("invalid url template: %w", err)
	}
	if len(templateArgs) == 0 {
		return nil, errors.New("url template produced no arguments")
	}
	args = append(args, templateArgs...)

	return args, nil
}

func splitCommandLine(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}

	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\n':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	if escaped {
		return nil, errors.New("dangling escape")
	}
	if quote != 0 {
		return nil, errors.New("unclosed quote")
	}
	flush()
	return out, nil
}

func joinExecArgs(args []string) string {
	escaped := make([]string, 0, len(args))
	for _, arg := range args {
		escaped = append(escaped, escapeExecArg(arg))
	}
	return strings.Join(escaped, " ")
}

func escapeDesktopValue(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func escapeExecArg(s string) string {
	if s == "" {
		return `""`
	}

	s = strings.ReplaceAll(s, "%", "%%")

	needsQuotes := false
	for _, r := range s {
		switch r {
		case ' ', '\t', '\n', '"', '\'', '\\', '&', '?', '#', ';', '(', ')', '<', '>', '|', '~', '*', '[', ']', '{', '}', '$', '`', '%':
			needsQuotes = true
		}
		if needsQuotes {
			break
		}
	}
	if !needsQuotes {
		return s
	}

	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\"", "\\\"",
		"$", "\\\\$",
		"`", "\\`",
	)
	return `"` + replacer.Replace(s) + `"`
}
