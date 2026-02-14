package tui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dopejs/opencc/internal/config"
)

// AppScreen represents the current screen in the new TUI.
type AppScreen int

const (
	ScreenMenu AppScreen = iota
	ScreenDashboard
	ScreenSettings
	ScreenProviderEdit
	ScreenProfileEdit
	ScreenLaunch
)

// NewAppModel is the main application model for the new TUI.
type NewAppModel struct {
	screen    AppScreen
	menu      MenuModel
	dashboard DashboardModel
	settings  SettingsModel
	editor    editorModel
	fallback  fallbackModel
	launch    LaunchModel
	width     int
	height    int
}

// NewNewAppModel creates a new application model.
func NewNewAppModel() NewAppModel {
	return NewAppModel{
		screen:    ScreenMenu,
		menu:      NewMenuModel(),
		dashboard: NewDashboardModel(),
		settings:  NewSettingsModel(),
		launch:    NewLaunchModel(),
	}
}

// Init implements tea.Model.
func (m NewAppModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m NewAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.screen {
	case ScreenMenu:
		return m.updateMenu(msg)
	case ScreenDashboard:
		return m.updateDashboard(msg)
	case ScreenSettings:
		return m.updateSettings(msg)
	case ScreenProviderEdit:
		return m.updateProviderEdit(msg)
	case ScreenProfileEdit:
		return m.updateProfileEdit(msg)
	case ScreenLaunch:
		return m.updateLaunch(msg)
	}

	return m, nil
}

func (m NewAppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case MenuSelectedMsg:
		switch msg.Action {
		case MenuLaunch:
			m.screen = ScreenLaunch
			m.launch.Refresh()
			return m, nil
		case MenuConfigure:
			m.screen = ScreenDashboard
			m.dashboard.Refresh()
			return m, nil
		case MenuSettings:
			m.screen = ScreenSettings
			m.settings.Refresh()
			return m, m.settings.Init()
		case MenuWebUI:
			// Open web UI in browser
			port := config.GetWebPort()
			url := fmt.Sprintf("http://127.0.0.1:%d", port)
			openBrowser(url)
			return m, nil
		case MenuQuit:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m NewAppModel) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DashboardBackMsg:
		m.screen = ScreenMenu
		m.menu.Refresh()
		return m, nil
	case DashboardEditProviderMsg:
		m.screen = ScreenProviderEdit
		m.editor = newEditorModel(msg.Name)
		return m, m.editor.init()
	case DashboardEditProfileMsg:
		m.screen = ScreenProfileEdit
		m.fallback = newFallbackModel(msg.Name)
		return m, m.fallback.init()
	case DashboardAddProviderMsg:
		m.screen = ScreenProviderEdit
		m.editor = newEditorModel("")
		return m, m.editor.init()
	case DashboardAddProfileMsg:
		// Create new profile with default name
		m.screen = ScreenProfileEdit
		m.fallback = newFallbackModel("")
		return m, m.fallback.init()
	}

	var cmd tea.Cmd
	m.dashboard, cmd = m.dashboard.Update(msg)
	return m, cmd
}

func (m NewAppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case SettingsSavedMsg, SettingsCancelledMsg:
		m.screen = ScreenMenu
		m.menu.Refresh()
		return m, nil
	}

	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)
	return m, cmd
}

func (m NewAppModel) updateProviderEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages from editor
	switch msg.(type) {
	case switchToListMsg:
		m.screen = ScreenDashboard
		m.dashboard.Refresh()
		return m, nil
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.update(msg)
	return m, cmd
}

func (m NewAppModel) updateProfileEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages from fallback editor
	switch msg.(type) {
	case switchToListMsg:
		m.screen = ScreenDashboard
		m.dashboard.Refresh()
		return m, nil
	}

	var cmd tea.Cmd
	m.fallback, cmd = m.fallback.update(msg)
	return m, cmd
}

func (m NewAppModel) updateLaunch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LaunchBackMsg:
		m.screen = ScreenMenu
		return m, nil
	case LaunchStartMsg:
		// Return the launch command to be executed
		return m, tea.Batch(
			tea.Quit,
			func() tea.Msg { return msg },
		)
	}

	var cmd tea.Cmd
	m.launch, cmd = m.launch.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m NewAppModel) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.menu.View()
	case ScreenDashboard:
		return m.dashboard.View()
	case ScreenSettings:
		return m.settings.View()
	case ScreenProviderEdit:
		return m.editor.view(m.width, m.height)
	case ScreenProfileEdit:
		return m.fallback.view(m.width, m.height)
	case ScreenLaunch:
		return m.launch.View()
	}
	return ""
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	}
}

// LaunchResult holds the result of the launch wizard.
type LaunchResult struct {
	Profile string
	CLI     string
}

// RunNewApp runs the new TUI application.
// Returns LaunchResult if user selected Launch, nil otherwise.
func RunNewApp() (*LaunchResult, error) {
	p := tea.NewProgram(NewNewAppModel(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	// Check if we got a launch message
	if m, ok := finalModel.(NewAppModel); ok {
		if m.screen == ScreenLaunch {
			return &LaunchResult{
				Profile: m.launch.selectedProfile,
				CLI:     m.launch.selectedCLI,
			}, nil
		}
	}

	return nil, nil
}
