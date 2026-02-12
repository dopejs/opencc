package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

type pickModel struct {
	choices   []string // all available providers
	order     []string // selected providers in order
	cursor    int
	grabbed   bool
	done      bool
	cancelled bool
}

func newPickModel() pickModel {
	names := config.ProviderNames()
	return pickModel{
		choices: names,
	}
}

func (m pickModel) Init() tea.Cmd {
	return nil
}

// orderIndex returns the 1-based index in order, or 0 if not selected.
func (m pickModel) orderIndex(name string) int {
	for i, n := range m.order {
		if n == name {
			return i + 1
		}
	}
	return 0
}

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.grabbed {
		return m.updateGrabbed(msg)
	}

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
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.choices) {
				name := m.choices[m.cursor]
				if idx := m.orderIndex(name); idx > 0 {
					m.order = removeFromOrder(m.order, name)
				} else {
					m.order = append(m.order, name)
				}
			}
		case "enter":
			if m.cursor < len(m.choices) {
				name := m.choices[m.cursor]
				if m.orderIndex(name) > 0 {
					m.grabbed = true
					return m, nil
				}
			}
			m.done = true
			return m, tea.Quit
		case "ctrl+s":
			if !isMac {
				m.done = true
				return m, tea.Quit
			}
		case "cmd+s":
			if isMac {
				m.done = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m pickModel) updateGrabbed(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.choices) {
		m.grabbed = false
		return m, nil
	}
	name := m.choices[m.cursor]
	orderIdx := m.orderIndex(name)
	if orderIdx == 0 {
		m.grabbed = false
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			m.grabbed = false
		case "up", "k":
			if orderIdx > 1 {
				m.order[orderIdx-1], m.order[orderIdx-2] = m.order[orderIdx-2], m.order[orderIdx-1]
			}
		case "down", "j":
			if orderIdx < len(m.order) {
				m.order[orderIdx-1], m.order[orderIdx] = m.order[orderIdx], m.order[orderIdx-1]
			}
		}
	}
	return m, nil
}

func (m pickModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("🎯 Select Providers")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Content box
	var content strings.Builder
	content.WriteString(sectionTitleStyle.Render(" Choose providers for this session"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render(" Space to toggle, Enter to reorder"))
	content.WriteString("\n\n")

	for i, name := range m.choices {
		cursor := "  "
		style := tableRowStyle
		orderIdx := m.orderIndex(name)

		if i == m.cursor {
			if m.grabbed {
				cursor = "⇕ "
				style = grabbedStyle
			} else {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}
		}

		var checkbox string
		if orderIdx > 0 {
			checkbox = lipgloss.NewStyle().
				Foreground(successColor).
				Render(fmt.Sprintf("[%d]", orderIdx))
		} else {
			checkbox = dimStyle.Render("[ ]")
		}

		content.WriteString(style.Render(cursor + checkbox + " " + name))
		if i < len(m.choices)-1 {
			content.WriteString("\n")
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(40).
		Render(content.String())
	b.WriteString(contentBox)

	b.WriteString("\n\n")
	if m.grabbed {
		b.WriteString(helpStyle.Render("  ↑↓ reorder • enter/esc drop"))
	} else {
		b.WriteString(helpStyle.Render("  space toggle • enter reorder/confirm • " + saveKeyHint() + " confirm • q cancel"))
	}

	return b.String()
}

// Result returns the selected provider names in order.
func (m pickModel) Result() []string {
	return m.order
}
