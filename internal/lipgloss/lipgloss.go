package lipgloss

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

type Color string

type Position int

const (
	Left Position = iota
	Right
	Top
)

type Style struct {
	width        int
	align        Position
	marginLeft   int
	paddingLeft  int
	paddingRight int
	setString    string
	foreground   Color
	faint        bool
	bold         bool
}

var (
	ansiRE        = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	colorEnabled  bool
	colorInitOnce sync.Once
)

func NewStyle() Style                    { return Style{align: Left} }
func (s Style) Foreground(c Color) Style { s.foreground = c; return s }
func (s Style) Faint(v bool) Style       { s.faint = v; return s }
func (s Style) Bold(v bool) Style        { s.bold = v; return s }
func (s Style) Width(w int) Style        { s.width = w; return s }
func (s Style) Align(p Position) Style   { s.align = p; return s }
func (s Style) MarginLeft(n int) Style   { s.marginLeft = n; return s }
func (s Style) PaddingLeft(n int) Style  { s.paddingLeft = n; return s }
func (s Style) PaddingRight(n int) Style { s.paddingRight = n; return s }

func (s Style) Padding(_ int, horizontal int) Style {
	s.paddingLeft = horizontal
	s.paddingRight = horizontal
	return s
}

func (s Style) SetString(v string) Style { s.setString = v; return s }

func (s Style) Render(parts ...string) string {
	out := strings.Join(parts, "")
	if out == "" {
		out = s.setString
	}

	if s.width > 0 {
		w := Width(out)
		if w < s.width {
			pad := strings.Repeat(" ", s.width-w)
			switch s.align {
			case Right:
				out = pad + out
			default:
				out = out + pad
			}
		}
	}

	if s.paddingLeft > 0 {
		out = strings.Repeat(" ", s.paddingLeft) + out
	}
	if s.paddingRight > 0 {
		out = out + strings.Repeat(" ", s.paddingRight)
	}
	if s.marginLeft > 0 {
		out = strings.Repeat(" ", s.marginLeft) + out
	}

	prefix := s.ansiPrefix()
	if prefix == "" || out == "" {
		return out
	}
	return prefix + out + "\x1b[0m"
}

func (s Style) ansiPrefix() string {
	if !colorsEnabled() {
		return ""
	}

	codes := make([]string, 0, 3)
	if s.bold {
		codes = append(codes, "1")
	}
	if s.faint {
		codes = append(codes, "2")
	}
	if fg := ansiForeground(s.foreground); fg != "" {
		codes = append(codes, fg)
	}
	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

func JoinHorizontal(_ Position, parts ...string) string { return strings.Join(parts, "") }

func Width(s string) int {
	clean := ansiRE.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}

func colorsEnabled() bool {
	colorInitOnce.Do(func() {
		term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
		if os.Getenv("NO_COLOR") != "" || term == "dumb" {
			colorEnabled = false
			return
		}
		colorEnabled = true
	})
	return colorEnabled
}

func ansiForeground(c Color) string {
	r, g, b, ok := parseHexColor(string(c))
	if !ok {
		return ""
	}
	return "38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b)
}

func parseHexColor(s string) (int, int, int, bool) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	switch len(s) {
	case 3:
		s = strings.Repeat(string(s[0]), 2) + strings.Repeat(string(s[1]), 2) + strings.Repeat(string(s[2]), 2)
	case 6:
	case 8:
		s = s[:6]
	default:
		return 0, 0, 0, false
	}

	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return int(r), int(g), int(b), true
}
