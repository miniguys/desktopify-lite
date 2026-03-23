// styles.go
package app

import (
	"github.com/miniguys/desktopify-lite/internal/lipgloss"
)

var (
	//
	// Colors
	styleGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	styleGreen1 = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5ab504"))

	//
	// Status
	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Padding(0, 1).
			SetString("✔")

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			SetString("✘")

	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F2BD38")).
			Padding(0, 1).
			SetString("✘")

	styleProcess = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("#f0f0f033"))

	//
	// Logo parts
	styleBorder = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1cb6be")).
		// PaddingBottom(1).
		PaddingLeft(1)
		// PaddingRight(1).
		// MarginBottom(0).
		// MarginLeft(1).
		// Border(lipgloss.ThickBorder()).
		// BorderForeground(lipgloss.Color("#296b31"))

	styleLogoCredits = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#484848")).
				Faint(true)

	//
	// Elements
	styleInputs = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#42B7C9"))

	styleHints = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d3d3d3")).
			Faint(true)

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ADD8")).
			Bold(true).
			MarginLeft(1)

	styleInfoName = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7a7a7a")).
			Faint(true)

	styleInfoValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Faint(true)

	styleQuestionContainer = lipgloss.NewStyle().
				Width(60).
				Align(lipgloss.Left)
)
