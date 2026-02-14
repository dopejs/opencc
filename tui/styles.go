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

// Layout constants - use at least 80% of terminal width
const (
	minContentWidth  = 80  // Minimum content width
	maxContentWidth  = 160 // Maximum content width
	contentWidthPct  = 85  // Percentage of terminal width to use
	horizontalMargin = 2   // Margin from terminal edges
	verticalMargin   = 1   // Margin from top/bottom
)

// LayoutDimensions calculates the content area dimensions based on terminal size.
// Returns contentWidth, contentHeight, leftPadding, topPadding
func LayoutDimensions(termWidth, termHeight int) (int, int, int, int) {
	// Calculate content width as percentage of terminal
	contentWidth := termWidth * contentWidthPct / 100

	// Apply min/max constraints
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth
	}
	if contentWidth > maxContentWidth {
		contentWidth = maxContentWidth
	}

	// Don't exceed terminal width minus margins
	if contentWidth > termWidth-horizontalMargin*2 {
		contentWidth = termWidth - horizontalMargin*2
	}

	// Calculate centering padding
	leftPadding := (termWidth - contentWidth) / 2
	if leftPadding < horizontalMargin {
		leftPadding = horizontalMargin
	}

	// Content height
	contentHeight := termHeight - verticalMargin*2 - 2 // -2 for help bar

	return contentWidth, contentHeight, leftPadding, verticalMargin
}

// WrapWithLayout wraps content with proper centering and padding
func WrapWithLayout(content string, termWidth, termHeight int) string {
	contentWidth, _, leftPadding, topPadding := LayoutDimensions(termWidth, termHeight)

	wrapper := lipgloss.NewStyle().
		Width(contentWidth).
		PaddingLeft(leftPadding).
		PaddingTop(topPadding)

	return wrapper.Render(content)
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

// RenderHelpBar renders a full-width help bar at the bottom of the screen.
// The help bar has a background color and left padding of 2 characters.
func RenderHelpBar(text string, termWidth int) string {
	helpBarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")). // dark gray background
		Foreground(dimColor).
		PaddingLeft(2).
		Width(termWidth)

	return helpBarStyle.Render(text)
}
