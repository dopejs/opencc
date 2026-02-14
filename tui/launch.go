package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

// LaunchBackMsg is sent when user wants to go back from launch wizard.
type LaunchBackMsg struct{}

// LaunchStartMsg is sent when user confirms launch.
type LaunchStartMsg struct {
	Profile string
	CLI     string
}

// LaunchModel is the launch wizard screen.
type LaunchModel struct {
	profiles        []string
	clis            []string
	profileCursor   int
	cliCursor       int
	focusOnCLI      bool
	selectedProfile string
	selectedCLI     string
	width           int
	height          int

	// Styles
	titleStyle    lipgloss.Style
	labelStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	helpStyle     lipgloss.Style
	boxStyle      lipgloss.Style
}

// NewLaunchModel creates a new launch wizard.
func NewLaunchModel() LaunchModel {
	profiles := config.ListProfiles()
	defaultProfile := config.GetDefaultProfile()
	defaultCLI := config.GetDefaultCLI()

	// Find default profile index
	profileIdx := 0
	for i, p := range profiles {
		if p == defaultProfile {
			profileIdx = i
			break
		}
	}

	// Find default CLI index
	clis := []string{"claude", "codex", "opencode"}
	cliIdx := 0
	for i, c := range clis {
		if c == defaultCLI {
			cliIdx = i
			break
		}
	}

	return LaunchModel{
		profiles:        profiles,
		clis:            clis,
		profileCursor:   profileIdx,
		cliCursor:       cliIdx,
		selectedProfile: defaultProfile,
		selectedCLI:     defaultCLI,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			MarginBottom(1),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginBottom(1),
		itemStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("7")),
		selectedStyle: lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(lipgloss.Color("14")).
			Bold(true),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(2),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 2),
	}
}

// Init implements tea.Model.
func (m LaunchModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m LaunchModel) Update(msg tea.Msg) (LaunchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.focusOnCLI {
				if m.cliCursor > 0 {
					m.cliCursor--
				}
			} else {
				if m.profileCursor > 0 {
					m.profileCursor--
				}
			}
		case "down", "j":
			if m.focusOnCLI {
				if m.cliCursor < len(m.clis)-1 {
					m.cliCursor++
				}
			} else {
				if m.profileCursor < len(m.profiles)-1 {
					m.profileCursor++
				}
			}
		case "tab", "left", "right":
			m.focusOnCLI = !m.focusOnCLI
		case "enter", " ":
			m.selectedProfile = m.profiles[m.profileCursor]
			m.selectedCLI = m.clis[m.cliCursor]
			return m, func() tea.Msg {
				return LaunchStartMsg{
					Profile: m.selectedProfile,
					CLI:     m.selectedCLI,
				}
			}
		case "esc", "q":
			return m, func() tea.Msg { return LaunchBackMsg{} }
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m LaunchModel) View() string {
	title := m.titleStyle.Render("Launch")

	// Profile list
	profileLabel := m.labelStyle.Render("Select Profile:")
	var profileList string
	for i, p := range m.profiles {
		line := p
		pc := config.GetProfileConfig(p)
		if pc != nil && len(pc.Providers) > 0 {
			line += fmt.Sprintf(" (%d providers)", len(pc.Providers))
		}

		if i == m.profileCursor {
			if !m.focusOnCLI {
				profileList += m.selectedStyle.Render("> "+line) + "\n"
			} else {
				profileList += m.itemStyle.Render("* "+line) + "\n"
			}
		} else {
			profileList += m.itemStyle.Render("  "+line) + "\n"
		}
	}
	profileBox := m.boxStyle.Render(profileLabel + "\n" + profileList)

	// CLI list
	cliLabel := m.labelStyle.Render("Select CLI:")
	cliDescriptions := map[string]string{
		"claude":   "Claude Code (Anthropic)",
		"codex":    "Codex CLI (OpenAI)",
		"opencode": "OpenCode",
	}
	var cliList string
	for i, c := range m.clis {
		line := c
		if desc, ok := cliDescriptions[c]; ok {
			line += " - " + desc
		}

		if i == m.cliCursor {
			if m.focusOnCLI {
				cliList += m.selectedStyle.Render("> "+line) + "\n"
			} else {
				cliList += m.itemStyle.Render("* "+line) + "\n"
			}
		} else {
			cliList += m.itemStyle.Render("  "+line) + "\n"
		}
	}
	cliBox := m.boxStyle.Render(cliLabel + "\n" + cliList)

	// Help
	help := m.helpStyle.Render("[Tab] switch  [Enter] launch  [Esc] back")

	// Layout
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, profileBox, "  ", cliBox),
		help,
	)

	return content
}

// Refresh reloads profiles and CLIs.
func (m *LaunchModel) Refresh() {
	m.profiles = config.ListProfiles()
	defaultProfile := config.GetDefaultProfile()
	defaultCLI := config.GetDefaultCLI()

	// Reset cursors to defaults
	for i, p := range m.profiles {
		if p == defaultProfile {
			m.profileCursor = i
			break
		}
	}
	for i, c := range m.clis {
		if c == defaultCLI {
			m.cliCursor = i
			break
		}
	}

	m.selectedProfile = defaultProfile
	m.selectedCLI = defaultCLI
	m.focusOnCLI = false
}
