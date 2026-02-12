package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dopejs/opencc/internal/config"
)

// groupCreateModel handles standalone group creation.
type groupCreateModel struct {
	nameInput textinput.Model
	cancelled bool
	created   string
	err       string
}

func newGroupCreateModel() groupCreateModel {
	ti := textinput.New()
	ti.Placeholder = "group name"
	ti.Prompt = "  Name: "
	ti.CharLimit = 64
	ti.Focus()
	return groupCreateModel{nameInput: ti}
}

func (m groupCreateModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m groupCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			name := strings.TrimSpace(m.nameInput.Value())
			if name == "" {
				m.err = "name is required"
				return m, nil
			}
			// Check if group already exists
			existing := config.ListProfiles()
			for _, p := range existing {
				if p == name {
					m.err = fmt.Sprintf("group %q already exists", name)
					return m, nil
				}
			}
			// Create empty group
			if err := config.WriteProfileOrder(name, nil); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.created = name
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m groupCreateModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Create New Group"))
	b.WriteString("\n\n")
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render("  " + m.err))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("  enter:create  esc:cancel"))

	return b.String()
}

// RunGroupCreate runs a standalone group creation TUI.
// Returns the created group name.
func RunGroupCreate() (string, error) {
	m := newGroupCreateModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	gm := result.(groupCreateModel)
	if gm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return gm.created, nil
}
