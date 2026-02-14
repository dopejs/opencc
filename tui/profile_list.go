package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dopejs/opencc/internal/config"
)

type profileListModel struct {
	profiles  []profileEntry
	cursor    int
	status    string
	deleting  bool
	creating  bool
	nameInput textinput.Model
}

type profileEntry struct {
	name  string
	count int // number of providers
}

func newProfileListModel() profileListModel {
	ti := textinput.New()
	ti.Placeholder = "group name"
	ti.Prompt = "  Name: "
	ti.CharLimit = 64
	return profileListModel{nameInput: ti}
}

type profilesLoadedMsg struct {
	profiles []profileEntry
}

func (m profileListModel) init() tea.Cmd {
	return func() tea.Msg {
		names := config.ListProfiles()
		var entries []profileEntry
		for _, n := range names {
			order, _ := config.ReadProfileOrder(n)
			entries = append(entries, profileEntry{name: n, count: len(order)})
		}
		return profilesLoadedMsg{profiles: entries}
	}
}

func (m profileListModel) update(msg tea.Msg) (profileListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case profilesLoadedMsg:
		m.profiles = msg.profiles
		m.cursor = 0
		m.deleting = false
		m.creating = false
		return m, nil

	case tea.KeyMsg:
		if m.creating {
			return m.handleCreate(msg)
		}
		if m.deleting {
			return m.handleDeleteConfirm(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m profileListModel) handleKey(msg tea.KeyMsg) (profileListModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m, func() tea.Msg { return switchToListMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
		}
	case "enter", "e":
		if len(m.profiles) > 0 {
			profile := m.profiles[m.cursor].name
			return m, func() tea.Msg { return switchToFallbackMsg{profile: profile} }
		}
	case "a":
		m.creating = true
		m.nameInput.SetValue("")
		m.nameInput.Focus()
		return m, textinput.Blink
	case "d":
		if len(m.profiles) > 0 {
			defaultProfile := config.GetDefaultProfile()
			if m.profiles[m.cursor].name == defaultProfile {
				m.status = fmt.Sprintf("Cannot delete the default profile '%s'", defaultProfile)
			} else {
				m.deleting = true
			}
		}
	}
	return m, nil
}

func (m profileListModel) handleDeleteConfirm(msg tea.KeyMsg) (profileListModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.profiles) {
			name := m.profiles[m.cursor].name
			config.DeleteProfile(name)
			m.deleting = false
			return m, m.init()
		}
	case "n", "N", "esc":
		m.deleting = false
	}
	return m, nil
}

func (m profileListModel) handleCreate(msg tea.KeyMsg) (profileListModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.creating = false
		m.nameInput.Blur()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			return m, nil
		}
		m.creating = false
		m.nameInput.Blur()
		// Create empty group and enter its editor
		config.WriteProfileOrder(name, nil)
		return m, func() tea.Msg { return switchToFallbackMsg{profile: name} }
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m profileListModel) view(width, height int) string {
	sidePadding := 2
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Groups"))
	b.WriteString("\n\n")

	if len(m.profiles) == 0 {
		b.WriteString("  No groups found.\n")
		b.WriteString("  Press 'a' to create a new group.\n")
	} else {
		for i, p := range m.profiles {
			cursor := "  "
			style := dimStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}
			line := fmt.Sprintf("%s%-14s [%d providers]", cursor, p.name, p.count)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.creating {
		b.WriteString(m.nameInput.View())
	} else if m.deleting && m.cursor < len(m.profiles) {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete group '%s'? (y/n)", m.profiles[m.cursor].name)))
	} else {
		if m.status != "" {
			b.WriteString(errorStyle.Render("  " + m.status))
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
	var helpText string
	if m.creating {
		helpText = "Enter create • Esc cancel"
	} else {
		helpText = "Enter edit • a new • d delete • Esc back"
	}
	helpBar := RenderHelpBar(helpText, width)
	view.WriteString(helpBar)

	return view.String()
}
