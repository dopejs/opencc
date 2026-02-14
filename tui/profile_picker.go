package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dopejs/opencc/internal/config"
)

// profilePickerModel is a standalone TUI for selecting a fallback profile.
type profilePickerModel struct {
	profiles  []string
	cursor    int
	selected  string
	cancelled bool
}

func newProfilePickerModel() profilePickerModel {
	return profilePickerModel{
		profiles: config.ListProfiles(),
	}
}

func (m profilePickerModel) Init() tea.Cmd {
	return nil
}

func (m profilePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.profiles) > 0 {
				m.selected = m.profiles[m.cursor]
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m profilePickerModel) View() string {
	width := 80  // default width
	height := 24 // default height

	sidePadding := 2
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Select provider group:"))
	b.WriteString("\n\n")

	if len(m.profiles) == 0 {
		b.WriteString("  No profiles found.\n")
	} else {
		for i, name := range m.profiles {
			cursor := "  "
			style := dimStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}
			b.WriteString(style.Render(cursor + name))
			b.WriteString("\n")
		}
	}

	// Build view with side padding
	mainContent := b.String()
	var view strings.Builder
	lines := strings.Split(mainContent, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", sidePadding))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Fill remaining space to push help bar to bottom
	currentLines := len(lines)
	remainingLines := height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom
	helpBar := RenderHelpBar("Enter select • q cancel", width)
	view.WriteString(helpBar)

	return view.String()
}
