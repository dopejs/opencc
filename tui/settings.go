package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/tui/components"
)

// SettingsModel is the global settings screen.
type SettingsModel struct {
	form     components.FormModel
	width    int
	height   int
	profiles []string
	saved    bool
	err      string

	// Styles
	titleStyle  lipgloss.Style
	statusStyle lipgloss.Style
}

// SettingsSavedMsg is sent when settings are saved.
type SettingsSavedMsg struct{}

// SettingsCancelledMsg is sent when settings are cancelled.
type SettingsCancelledMsg struct{}

// NewSettingsModel creates a new settings screen.
func NewSettingsModel() SettingsModel {
	profiles := config.ListProfiles()
	currentProfile := config.GetDefaultProfile()
	currentCLI := config.GetDefaultCLI()
	currentPort := config.GetWebPort()

	fields := []components.Field{
		{
			Key:     "default_cli",
			Label:   "Default CLI",
			Type:    components.FieldSelect,
			Value:   currentCLI,
			Options: []string{"claude", "codex", "opencode"},
		},
		{
			Key:     "default_profile",
			Label:   "Default Profile",
			Type:    components.FieldSelect,
			Value:   currentProfile,
			Options: profiles,
		},
		{
			Key:         "web_port",
			Label:       "Web UI Port",
			Type:        components.FieldText,
			Value:       strconv.Itoa(currentPort),
			Placeholder: "19840",
		},
	}

	form := components.NewForm(fields)

	return SettingsModel{
		form:     form,
		profiles: profiles,
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			MarginBottom(1),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")),
	}
}

// Init implements tea.Model.
func (m SettingsModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.form.SetSize(msg.Width-4, msg.Height-10)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			return m, m.save()
		case "esc":
			return m, func() tea.Msg { return SettingsCancelledMsg{} }
		}
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m SettingsModel) save() tea.Cmd {
	values := m.form.GetValues()

	// Validate and save default CLI
	if cli := values["default_cli"]; cli != "" {
		if err := config.SetDefaultCLI(cli); err != nil {
			m.err = err.Error()
			return nil
		}
	}

	// Validate and save default profile
	if profile := values["default_profile"]; profile != "" {
		if err := config.SetDefaultProfile(profile); err != nil {
			m.err = err.Error()
			return nil
		}
	}

	// Validate and save web port
	if portStr := values["web_port"]; portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			m.err = "Invalid port number"
			return nil
		}
		if port < 1024 || port > 65535 {
			m.err = "Port must be between 1024 and 65535"
			return nil
		}
		if err := config.SetWebPort(port); err != nil {
			m.err = err.Error()
			return nil
		}
	}

	m.saved = true
	return func() tea.Msg { return SettingsSavedMsg{} }
}

// View implements tea.Model.
func (m SettingsModel) View() string {
	// Use global layout dimensions
	contentWidth, _, _, _ := LayoutDimensions(m.width, m.height)
	sidePadding := 2

	title := m.titleStyle.Render("Settings")

	formView := m.form.View()

	// Wrap form in a box with proper width
	formWidth := contentWidth * 70 / 100
	if formWidth < 50 {
		formWidth = 50
	}
	if formWidth > 80 {
		formWidth = 80
	}

	formBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(formWidth).
		Padding(1, 2).
		Render(formView)

	var status string
	if m.saved {
		status = m.statusStyle.Render("Settings saved!")
	}
	if m.err != "" {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: " + m.err)
	}

	mainContent := fmt.Sprintf("%s\n\n%s", title, formBox)
	if status != "" {
		mainContent += "\n\n" + status
	}

	// Build view with side padding
	var view strings.Builder
	lines := strings.Split(mainContent, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", sidePadding))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Fill remaining space to push help bar to bottom
	currentLines := len(lines)
	remainingLines := m.height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom
	helpBar := RenderHelpBar("Tab next • Shift+Tab prev • Ctrl+S save • Esc back", m.width)
	view.WriteString(helpBar)

	return view.String()
}

// Refresh reloads settings from config.
func (m *SettingsModel) Refresh() {
	m.profiles = config.ListProfiles()
	m.form.SetValue("default_cli", config.GetDefaultCLI())
	m.form.SetValue("default_profile", config.GetDefaultProfile())
	m.form.SetValue("web_port", strconv.Itoa(config.GetWebPort()))
	m.saved = false
	m.err = ""
}
