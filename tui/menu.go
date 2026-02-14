package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

// MenuAction represents a menu item action.
type MenuAction int

const (
	MenuLaunch MenuAction = iota
	MenuConfigure
	MenuSettings
	MenuWebUI
	MenuQuit
)

type menuItem struct {
	label  string
	action MenuAction
}

// MenuModel is the main menu screen.
type MenuModel struct {
	items   []menuItem
	cursor  int
	width   int
	height  int
	profile string
	cli     string

	// Styles
	titleStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	statusStyle   lipgloss.Style
	boxStyle      lipgloss.Style
}

// NewMenuModel creates a new main menu.
func NewMenuModel() MenuModel {
	return MenuModel{
		items: []menuItem{
			{label: "Launch", action: MenuLaunch},
			{label: "Configure", action: MenuConfigure},
			{label: "Settings", action: MenuSettings},
			{label: "Web UI", action: MenuWebUI},
			{label: "Quit", action: MenuQuit},
		},
		profile: config.GetDefaultProfile(),
		cli:     config.GetDefaultCLI(),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")),
		itemStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")),
		selectedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 4),
	}
}

// Init implements tea.Model.
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// MenuSelectedMsg is sent when a menu item is selected.
type MenuSelectedMsg struct {
	Action MenuAction
}

// Update implements tea.Model.
func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, func() tea.Msg {
				return MenuSelectedMsg{Action: m.items[m.cursor].action}
			}
		case "q", "esc":
			return m, tea.Quit
		case "1":
			m.cursor = 0
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuLaunch} }
		case "2":
			m.cursor = 1
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuConfigure} }
		case "3":
			m.cursor = 2
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuSettings} }
		case "4":
			m.cursor = 3
			return m, func() tea.Msg { return MenuSelectedMsg{Action: MenuWebUI} }
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m MenuModel) View() string {
	// Title and subtitle - centered
	title := m.titleStyle.Render("OpenCC")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render("Environment Switcher")

	// Menu items - fixed width, cursor doesn't shift text
	var menuItems strings.Builder
	for i, item := range m.items {
		if i == m.cursor {
			menuItems.WriteString(m.selectedStyle.Render("> " + item.label))
		} else {
			menuItems.WriteString(m.itemStyle.Render("  " + item.label))
		}
		menuItems.WriteString("\n")
	}

	// Status line - centered below menu
	status := m.statusStyle.Render(fmt.Sprintf("Profile: %s  |  CLI: %s", m.profile, m.cli))

	// Calculate box width
	boxWidth := 36

	// Build box content: title and subtitle centered, menu left-aligned
	titleCentered := lipgloss.NewStyle().Width(boxWidth - 10).Align(lipgloss.Center).Render(title)
	subtitleCentered := lipgloss.NewStyle().Width(boxWidth - 10).Align(lipgloss.Center).Render(subtitle)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleCentered,
		subtitleCentered,
		"",
		menuItems.String(),
	)

	box := m.boxStyle.Width(boxWidth).Render(content)

	// Center box on screen
	boxWidthActual := lipgloss.Width(box)
	boxHeight := lipgloss.Height(box)

	boxPadLeft := (m.width - boxWidthActual) / 2
	// Reserve 1 line for help bar at bottom
	boxPadTop := (m.height - boxHeight - 3) / 2

	if boxPadLeft < 0 {
		boxPadLeft = 0
	}
	if boxPadTop < 0 {
		boxPadTop = 0
	}

	// Build main content area
	var view strings.Builder
	for i := 0; i < boxPadTop; i++ {
		view.WriteString("\n")
	}

	lines := strings.Split(box, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", boxPadLeft))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Status below box
	view.WriteString(strings.Repeat(" ", boxPadLeft))
	view.WriteString(status)
	view.WriteString("\n")

	// Fill remaining space before help bar
	currentLines := strings.Count(view.String(), "\n")
	remainingLines := m.height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom - full terminal width with background
	helpBar := RenderHelpBar("↑↓ navigate • Enter select • q quit", m.width)
	view.WriteString(helpBar)

	return view.String()
}

// Refresh reloads config values.
func (m *MenuModel) Refresh() {
	m.profile = config.GetDefaultProfile()
	m.cli = config.GetDefaultCLI()
}
