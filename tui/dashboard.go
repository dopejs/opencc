package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/tui/components"
)

// DashboardModel is the main configuration dashboard with split view.
type DashboardModel struct {
	list        components.ListModel
	width       int
	height      int
	focusLeft   bool // true = sidebar focused, false = detail focused
	selectedID  string
	selectedType string // "provider", "profile", "binding"

	// Styles
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	labelStyle  lipgloss.Style
	valueStyle  lipgloss.Style
}

// DashboardBackMsg is sent when user wants to go back to menu.
type DashboardBackMsg struct{}

// DashboardEditProviderMsg is sent when user wants to edit a provider.
type DashboardEditProviderMsg struct {
	Name string
}

// DashboardEditProfileMsg is sent when user wants to edit a profile.
type DashboardEditProfileMsg struct {
	Name string
}

// DashboardAddProviderMsg is sent when user wants to add a provider.
type DashboardAddProviderMsg struct{}

// DashboardAddProfileMsg is sent when user wants to add a profile.
type DashboardAddProfileMsg struct{}

// NewDashboardModel creates a new dashboard.
func NewDashboardModel() DashboardModel {
	m := DashboardModel{
		focusLeft: true,
		borderStyle: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("240")),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")),
		labelStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
		valueStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
	}
	m.refreshList()
	return m
}

func (m *DashboardModel) refreshList() {
	providers := config.ProviderNames()
	profiles := config.ListProfiles()
	bindings := config.GetAllProjectBindings()

	var providerItems []components.ListItem
	for _, name := range providers {
		p := config.GetProvider(name)
		sublabel := ""
		if p != nil && p.BaseURL != "" {
			sublabel = p.BaseURL
		}
		providerItems = append(providerItems, components.ListItem{
			ID:       "provider:" + name,
			Label:    name,
			Sublabel: sublabel,
		})
	}

	var profileItems []components.ListItem
	defaultProfile := config.GetDefaultProfile()
	for _, name := range profiles {
		pc := config.GetProfileConfig(name)
		sublabel := ""
		if pc != nil {
			sublabel = fmt.Sprintf("%d providers", len(pc.Providers))
		}
		icon := ""
		if name == defaultProfile {
			icon = "★"
		}
		profileItems = append(profileItems, components.ListItem{
			ID:       "profile:" + name,
			Label:    name,
			Sublabel: sublabel,
			Icon:     icon,
		})
	}

	var bindingItems []components.ListItem
	for path, binding := range bindings {
		// Shorten path for display
		shortPath := path
		if len(shortPath) > 30 {
			shortPath = "..." + shortPath[len(shortPath)-27:]
		}
		// Build sublabel from binding
		var sublabel string
		if binding.Profile != "" && binding.CLI != "" {
			sublabel = "→ " + binding.Profile + " (" + binding.CLI + ")"
		} else if binding.Profile != "" {
			sublabel = "→ " + binding.Profile
		} else if binding.CLI != "" {
			sublabel = "→ (" + binding.CLI + ")"
		} else {
			sublabel = "→ (default)"
		}
		bindingItems = append(bindingItems, components.ListItem{
			ID:       "binding:" + path,
			Label:    shortPath,
			Sublabel: sublabel,
		})
	}

	sections := []components.ListSection{
		{Name: "Providers", Items: providerItems},
		{Name: "Profiles", Items: profileItems},
		{Name: "Project Bindings", Items: bindingItems, Collapsed: true},
	}

	m.list = components.NewList(sections)
	m.list.SetOnSelect(func(sectionIdx, itemIdx int, item components.ListItem) tea.Cmd {
		m.selectedID = item.ID
		parts := strings.SplitN(item.ID, ":", 2)
		if len(parts) == 2 {
			m.selectedType = parts[0]
		}
		return nil
	})
}

// Init implements tea.Model.
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate list size based on layout
		contentWidth, contentHeight, _, _ := LayoutDimensions(m.width, m.height)
		leftWidth := contentWidth * 35 / 100
		if leftWidth < 28 {
			leftWidth = 28
		}
		paneHeight := contentHeight - 2
		// List size accounts for border (2) and internal padding (2)
		m.list.SetSize(leftWidth-4, paneHeight-2)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return DashboardBackMsg{} }
		case "tab":
			m.focusLeft = !m.focusLeft
		case "a":
			// Add new item based on current section
			_, _, item, ok := m.list.GetSelectedItem()
			if ok {
				parts := strings.SplitN(item.ID, ":", 2)
				if len(parts) > 0 {
					switch parts[0] {
					case "provider":
						return m, func() tea.Msg { return DashboardAddProviderMsg{} }
					case "profile":
						return m, func() tea.Msg { return DashboardAddProfileMsg{} }
					}
				}
			}
		case "e", "enter":
			_, _, item, ok := m.list.GetSelectedItem()
			if ok {
				parts := strings.SplitN(item.ID, ":", 2)
				if len(parts) == 2 {
					switch parts[0] {
					case "provider":
						return m, func() tea.Msg { return DashboardEditProviderMsg{Name: parts[1]} }
					case "profile":
						return m, func() tea.Msg { return DashboardEditProfileMsg{Name: parts[1]} }
					}
				}
			}
		case "d":
			_, _, item, ok := m.list.GetSelectedItem()
			if ok {
				parts := strings.SplitN(item.ID, ":", 2)
				if len(parts) == 2 {
					switch parts[0] {
					case "provider":
						config.DeleteProviderByName(parts[1])
						m.refreshList()
					case "profile":
						if err := config.DeleteProfile(parts[1]); err != nil {
							// Can't delete default profile - ignore
						} else {
							m.refreshList()
						}
					case "binding":
						config.UnbindProject(parts[1])
						m.refreshList()
					}
				}
			}
		}
	}

	if m.focusLeft {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)

		// Update selection
		_, _, item, ok := m.list.GetSelectedItem()
		if ok {
			m.selectedID = item.ID
			parts := strings.SplitN(item.ID, ":", 2)
			if len(parts) == 2 {
				m.selectedType = parts[0]
			}
		}
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m DashboardModel) View() string {
	// Layout: 2 padding on each side
	sidePadding := 2
	contentWidth := m.width - sidePadding*2
	if contentWidth < 60 {
		contentWidth = 60
	}

	// Left pane takes 35%, right pane takes 65%
	leftWidth := contentWidth * 35 / 100
	rightWidth := contentWidth - leftWidth

	// Minimum widths
	if leftWidth < 28 {
		leftWidth = 28
	}
	if rightWidth < 40 {
		rightWidth = 40
	}

	// Pane height (reserve 1 for help bar)
	paneHeight := m.height - 1
	if paneHeight < 10 {
		paneHeight = 10
	}

	// Internal width for border and padding
	leftInternalWidth := leftWidth - 2 - 2 // border + padding
	rightInternalWidth := rightWidth - 2 - 2

	// Left pane - list
	leftContent := m.list.View()
	leftPane := m.borderStyle.
		Width(leftInternalWidth).
		Height(paneHeight - 2).
		Padding(0, 1).
		Render(leftContent)

	// Right pane - detail
	rightContent := m.renderDetail()
	rightPane := m.borderStyle.
		Width(rightInternalWidth).
		Height(paneHeight - 2).
		Padding(0, 1).
		Render(rightContent)

	// Join panes
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Build view with side padding
	var view strings.Builder
	lines := strings.Split(mainContent, "\n")
	for _, line := range lines {
		view.WriteString(strings.Repeat(" ", sidePadding))
		view.WriteString(line)
		view.WriteString("\n")
	}

	// Fill remaining space
	currentLines := len(lines)
	remainingLines := m.height - currentLines - 1
	for i := 0; i < remainingLines; i++ {
		view.WriteString("\n")
	}

	// Help bar at bottom
	helpBar := RenderHelpBar("a add • e edit • d delete • Tab switch pane • Esc back", m.width)
	view.WriteString(helpBar)

	return view.String()
}

func (m DashboardModel) renderDetail() string {
	if m.selectedID == "" {
		return m.labelStyle.Render("Select an item to view details")
	}

	parts := strings.SplitN(m.selectedID, ":", 2)
	if len(parts) != 2 {
		return ""
	}

	itemType := parts[0]
	itemName := parts[1]

	var b strings.Builder

	switch itemType {
	case "provider":
		p := config.GetProvider(itemName)
		if p == nil {
			return m.labelStyle.Render("Provider not found")
		}

		b.WriteString(m.titleStyle.Render("Provider: " + itemName))
		b.WriteString("\n\n")

		b.WriteString(m.labelStyle.Render("Base URL: "))
		b.WriteString(m.valueStyle.Render(p.BaseURL))
		b.WriteString("\n")

		b.WriteString(m.labelStyle.Render("Auth: "))
		if len(p.AuthToken) > 8 {
			b.WriteString(m.valueStyle.Render(p.AuthToken[:5] + "..." + p.AuthToken[len(p.AuthToken)-4:]))
		} else {
			b.WriteString(m.valueStyle.Render("****"))
		}
		b.WriteString("\n\n")

		b.WriteString(m.labelStyle.Render("Models:"))
		b.WriteString("\n")
		if p.Model != "" {
			b.WriteString("  Default: " + p.Model + "\n")
		}
		if p.ReasoningModel != "" {
			b.WriteString("  Reasoning: " + p.ReasoningModel + "\n")
		}
		if p.HaikuModel != "" {
			b.WriteString("  Haiku: " + p.HaikuModel + "\n")
		}
		if p.OpusModel != "" {
			b.WriteString("  Opus: " + p.OpusModel + "\n")
		}
		if p.SonnetModel != "" {
			b.WriteString("  Sonnet: " + p.SonnetModel + "\n")
		}

		if len(p.EnvVars) > 0 {
			b.WriteString("\n")
			b.WriteString(m.labelStyle.Render(fmt.Sprintf("Env Vars: %d configured", len(p.EnvVars))))
		}

		// Show which profiles use this provider
		profiles := config.ListProfiles()
		var usedIn []string
		for _, profile := range profiles {
			pc := config.GetProfileConfig(profile)
			if pc != nil {
				for i, prov := range pc.Providers {
					if prov == itemName {
						pos := "fallback"
						if i == 0 {
							pos = "primary"
						}
						usedIn = append(usedIn, fmt.Sprintf("%s (%s)", profile, pos))
						break
					}
				}
			}
		}
		if len(usedIn) > 0 {
			b.WriteString("\n\n")
			b.WriteString(m.labelStyle.Render("Used in profiles:"))
			b.WriteString("\n")
			for _, u := range usedIn {
				b.WriteString("  " + u + "\n")
			}
		}

	case "profile":
		pc := config.GetProfileConfig(itemName)
		if pc == nil {
			return m.labelStyle.Render("Profile not found")
		}

		defaultProfile := config.GetDefaultProfile()
		title := "Profile: " + itemName
		if itemName == defaultProfile {
			title += " ★"
		}
		b.WriteString(m.titleStyle.Render(title))
		b.WriteString("\n\n")

		b.WriteString(m.labelStyle.Render("Providers:"))
		b.WriteString("\n")
		for i, prov := range pc.Providers {
			pos := "fallback"
			if i == 0 {
				pos = "primary"
			}
			b.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, prov, pos))
		}

		if pc.LongContextThreshold > 0 {
			b.WriteString("\n")
			b.WriteString(m.labelStyle.Render(fmt.Sprintf("Long Context Threshold: %d tokens", pc.LongContextThreshold)))
		}

		if len(pc.Routing) > 0 {
			b.WriteString("\n\n")
			b.WriteString(m.labelStyle.Render("Scenario Routing:"))
			b.WriteString("\n")
			for scenario, route := range pc.Routing {
				if len(route.Providers) > 0 {
					pr := route.Providers[0]
					model := pr.Model
					if model == "" {
						model = "(default)"
					}
					b.WriteString(fmt.Sprintf("  %s → %s: %s\n", scenario, pr.Name, model))
				}
			}
		}

	case "binding":
		binding := config.GetProjectBinding(itemName)
		b.WriteString(m.titleStyle.Render("Project Binding"))
		b.WriteString("\n\n")
		b.WriteString(m.labelStyle.Render("Path: "))
		b.WriteString(m.valueStyle.Render(itemName))
		b.WriteString("\n")
		if binding != nil {
			profileVal := binding.Profile
			if profileVal == "" {
				profileVal = "(default)"
			}
			cliVal := binding.CLI
			if cliVal == "" {
				cliVal = "(default)"
			}
			b.WriteString(m.labelStyle.Render("Profile: "))
			b.WriteString(m.valueStyle.Render(profileVal))
			b.WriteString("\n")
			b.WriteString(m.labelStyle.Render("CLI: "))
			b.WriteString(m.valueStyle.Render(cliVal))
		}
	}

	return b.String()
}

// Refresh reloads the dashboard data.
func (m *DashboardModel) Refresh() {
	m.refreshList()
}
