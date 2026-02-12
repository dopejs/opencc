package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

type fallbackModel struct {
	profile    string   // profile name ("default", "work", etc.)
	allConfigs []string // all available providers
	order      []string // current fallback order (selected providers)
	cursor     int      // cursor position in allConfigs
	grabbed    bool     // true = item is grabbed and arrow keys reorder
	status     string
	saved      bool     // true = save succeeded, waiting to exit
}

func newFallbackModel(profile string) fallbackModel {
	if profile == "" {
		profile = "default"
	}
	return fallbackModel{profile: profile}
}

type fallbackLoadedMsg struct {
	allConfigs []string
	order      []string
}

func (m fallbackModel) init() tea.Cmd {
	profile := m.profile
	return func() tea.Msg {
		names := config.ProviderNames()
		order, _ := config.ReadProfileOrder(profile)
		return fallbackLoadedMsg{allConfigs: names, order: order}
	}
}

func (m fallbackModel) update(msg tea.Msg) (fallbackModel, tea.Cmd) {
	// After save, ignore everything except saveExitMsg
	if m.saved {
		if _, ok := msg.(saveExitMsg); ok {
			return m, func() tea.Msg { return switchToListMsg{} }
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case fallbackLoadedMsg:
		m.allConfigs = msg.allConfigs
		m.order = msg.order
		m.cursor = 0
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// orderIndex returns the 1-based index in order, or 0 if not in order.
func (m fallbackModel) orderIndex(name string) int {
	for i, n := range m.order {
		if n == name {
			return i + 1
		}
	}
	return 0
}

func (m fallbackModel) handleKey(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	if m.grabbed {
		return m.handleGrabbed(msg)
	}

	switch msg.String() {
	case "esc", "q":
		// Cancel — return without saving
		return m, func() tea.Msg { return switchToListMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.allConfigs)-1 {
			m.cursor++
		}
	case " ":
		// Toggle selection
		if m.cursor < len(m.allConfigs) {
			name := m.allConfigs[m.cursor]
			if idx := m.orderIndex(name); idx > 0 {
				// Remove from order
				m.order = removeFromOrder(m.order, name)
			} else {
				// Add to end of order
				m.order = append(m.order, name)
			}
		}
	case "enter":
		// Enter grab mode only if current item is in order
		if m.cursor < len(m.allConfigs) {
			name := m.allConfigs[m.cursor]
			if m.orderIndex(name) > 0 {
				m.grabbed = true
			}
		}
	case "s", "ctrl+s", "cmd+s":
		return m.saveAndExit()
	}
	return m, nil
}

func (m fallbackModel) saveAndExit() (fallbackModel, tea.Cmd) {
	if err := config.WriteProfileOrder(m.profile, m.order); err != nil {
		m.status = "Error: " + err.Error()
		return m, nil
	}
	m.saved = true
	m.status = "Saved"
	return m, saveExitTick()
}

func (m fallbackModel) handleGrabbed(msg tea.KeyMsg) (fallbackModel, tea.Cmd) {
	if m.cursor >= len(m.allConfigs) {
		m.grabbed = false
		return m, nil
	}
	name := m.allConfigs[m.cursor]
	orderIdx := m.orderIndex(name)
	if orderIdx == 0 {
		m.grabbed = false
		return m, nil
	}

	switch msg.String() {
	case "esc", "enter":
		m.grabbed = false
	case "up", "k":
		// Move up in order (swap with previous in order)
		if orderIdx > 1 {
			m.order[orderIdx-1], m.order[orderIdx-2] = m.order[orderIdx-2], m.order[orderIdx-1]
		}
	case "down", "j":
		// Move down in order (swap with next in order)
		if orderIdx < len(m.order) {
			m.order[orderIdx-1], m.order[orderIdx] = m.order[orderIdx], m.order[orderIdx-1]
		}
	}
	return m, nil
}

func removeFromOrder(order []string, name string) []string {
	var result []string
	for _, n := range order {
		if n != name {
			result = append(result, n)
		}
	}
	return result
}

func (m fallbackModel) view(width, height int) string {
	var b strings.Builder

	// Header
	title := "Group: default"
	if m.profile != "" && m.profile != "default" {
		title = fmt.Sprintf("Group: %s", m.profile)
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("📦 " + title)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Content box
	var content strings.Builder
	if len(m.allConfigs) == 0 {
		content.WriteString(dimStyle.Render("No providers configured.\n"))
		content.WriteString(dimStyle.Render("Run 'opencc config add provider' to create one."))
	} else {
		content.WriteString(sectionTitleStyle.Render(" Select Providers"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Space to toggle, Enter to reorder"))
		content.WriteString("\n\n")

		for i, name := range m.allConfigs {
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

			// Show order number or empty checkbox
			var checkbox string
			if orderIdx > 0 {
				checkbox = lipgloss.NewStyle().
					Foreground(successColor).
					Render(fmt.Sprintf("[%d]", orderIdx))
			} else {
				checkbox = dimStyle.Render("[ ]")
			}

			line := fmt.Sprintf("%s%s %s", cursor, checkbox, name)
			content.WriteString(style.Render(line))
			if i < len(m.allConfigs)-1 {
				content.WriteString("\n")
			}
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
	if m.saved {
		b.WriteString(successStyle.Render("  ✓ " + m.status))
	} else {
		if m.status != "" {
			statusBox := lipgloss.NewStyle().
				Foreground(errorColor).
				Render("  ✗ " + m.status)
			b.WriteString(statusBox)
			b.WriteString("\n")
		}
		if m.grabbed {
			b.WriteString(helpStyle.Render("  ↑↓ reorder • enter/esc drop"))
		} else {
			b.WriteString(helpStyle.Render("  ↑↓ move • space toggle • enter reorder • s/" + saveKeyHint() + " save • esc cancel"))
		}
	}

	return b.String()
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// standaloneFallbackModel wraps fallbackModel for standalone use.
type standaloneFallbackModel struct {
	fallback  fallbackModel
	cancelled bool
}

func newStandaloneFallbackModel(profile string) standaloneFallbackModel {
	return standaloneFallbackModel{
		fallback: newFallbackModel(profile),
	}
}

func (m standaloneFallbackModel) Init() tea.Cmd {
	return m.fallback.init()
}

func (m standaloneFallbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	case switchToListMsg:
		// Fallback editor finished — quit
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.fallback, cmd = m.fallback.update(msg)
	return m, cmd
}

func (m standaloneFallbackModel) View() string {
	return m.fallback.view(0, 0)
}

// RunEditProfile runs a standalone fallback editor TUI for editing a profile.
func RunEditProfile(name string) error {
	m := newStandaloneFallbackModel(name)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	sm := result.(standaloneFallbackModel)
	if sm.cancelled {
		return fmt.Errorf("cancelled")
	}
	return nil
}
