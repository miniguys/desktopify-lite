package app

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/miniguys/desktopify-lite/internal/lipgloss"
)

type KV struct {
	K string
	V string
}

var (
	version   = defaultVersion()
	commit    = "unknown"
	buildDate = "unknown"
	author    = "miniguys"
	infoBlock string
)

const infoLeftPad = 2

func defaultVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := strings.TrimSpace(info.Main.Version); v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

func init() {
	items := []KV{
		{K: "Author", V: author},
		{K: "Version", V: version},
		{K: "Commit", V: commit},
		{K: "BuildDate", V: buildDate},
		{
			K: "Description",
			V: "Generates .desktop file for selected\nwebsite.\nCan be configured with config file.",
		},
	}

	infoBlock = "\n" + renderKVBlock(items, styleInfoName, styleInfoValue, infoLeftPad)
}

func printInfoBlock() {
	fmt.Print(infoBlock)
	fmt.Println()
}

func versionLine() string {
	return fmt.Sprintf("desktopify-lite %s", version)
}

func printVersion() {
	fmt.Println(versionLine())
}

func renderKVBlock(items []KV, keyStyle, valStyle lipgloss.Style, leftPad int) string {
	pad := strings.Repeat(" ", max(0, leftPad))

	keyW := 0
	for _, it := range items {
		if w := lipgloss.Width(it.K); w > keyW {
			keyW = w
		}
	}

	keyCol := keyStyle.Width(keyW).Align(lipgloss.Right)
	sep := lipgloss.NewStyle().Render(": ")
	contIndent := pad + strings.Repeat(" ", keyW+lipgloss.Width(sep))

	var out []string
	for _, it := range items {
		lines := splitLinesKeepAtLeastOne(it.V)

		out = append(out,
			pad+lipgloss.JoinHorizontal(lipgloss.Top,
				keyCol.Render(it.K),
				sep,
				valStyle.Render(lines[0]),
			),
		)

		for _, ln := range lines[1:] {
			out = append(out, contIndent+valStyle.Render(ln))
		}
	}

	return strings.Join(out, "\n")
}

func splitLinesKeepAtLeastOne(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}
