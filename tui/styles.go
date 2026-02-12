package tui

import (
	"runtime"

	"github.com/charmbracelet/lipgloss"
)

// Platform detection
var isMac = runtime.GOOS == "darwin"

// SaveKey returns the appropriate save key hint for the current platform
func saveKeyHint() string {
	if isMac {
		return "⌘+S"
	}
	return "ctrl+s"
}

// Colors - soft, muted palette
var (
	primaryColor   = lipgloss.Color("109") // soft teal
	accentColor    = lipgloss.Color("146") // soft lavender
	successColor   = lipgloss.Color("108") // soft sage green
	errorColor     = lipgloss.Color("174") // soft coral
	dimColor       = lipgloss.Color("245") // light gray
	borderColor    = lipgloss.Color("240") // subtle gray
	highlightColor = lipgloss.Color("152") // soft mint
	headerBgColor  = lipgloss.Color("238") // dark gray bg
)

// Base styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	grabbedStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				MarginBottom(1)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(dimColor)

	tableRowStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	tableSelectedRowStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	// Badge styles
	badgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(primaryColor).
			Padding(0, 1)

	badgeDimStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			Background(headerBgColor).
			Padding(0, 1)
)

// Helper to create a bordered section
func renderSection(title string, content string) string {
	titleRendered := sectionTitleStyle.Render("┌─ " + title + " ")
	box := lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top:         "",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "",
			TopRight:    "",
			BottomLeft:  "└",
			BottomRight: "┘",
		}).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(content)
	return titleRendered + "\n" + box
}

// Helper to render a simple horizontal line
func renderLine(width int) string {
	return dimStyle.Render(lipgloss.NewStyle().
		Foreground(borderColor).
		Render("─" + repeatString("─", width-2) + "─"))
}

func repeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
