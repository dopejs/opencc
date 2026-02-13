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

	// Routing section
	section         int                                 // 0=default providers, 1=routing scenarios
	routingCursor   int                                 // cursor in routing scenarios
	routingExpanded map[config.Scenario]bool            // which scenarios are expanded
	routingOrder    map[config.Scenario][]string        // provider order per scenario
	routingModels   map[config.Scenario]map[string]string // per-provider models per scenario

	status string
	saved  bool // true = save succeeded, waiting to exit
}

func newFallbackModel(profile string) fallbackModel {
	if profile == "" {
		profile = "default"
	}
	return fallbackModel{
		profile:         profile,
		routingExpanded: make(map[config.Scenario]bool),
		routingOrder:    make(map[config.Scenario][]string),
		routingModels:   make(map[config.Scenario]map[string]string),
	}
}

type fallbackLoadedMsg struct {
	allConfigs []string
	order      []string
	routing    map[config.Scenario]*config.ScenarioRoute
}

func (m fallbackModel) init() tea.Cmd {
	profile := m.profile
	return func() tea.Msg {
		names := config.ProviderNames()
		pc := config.GetProfileConfig(profile)
		var order []string
		var routing map[config.Scenario]*config.ScenarioRoute
		if pc != nil {
			order = pc.Providers
			routing = pc.Routing
		}
		return fallbackLoadedMsg{allConfigs: names, order: order, routing: routing}
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
		// Load routing data
		if msg.routing != nil {
			for scenario, route := range msg.routing {
				m.routingOrder[scenario] = route.ProviderNames()
				m.routingModels[scenario] = make(map[string]string)
				for _, pr := range route.Providers {
					if pr.Model != "" {
						m.routingModels[scenario][pr.Name] = pr.Model
					}
				}
			}
		}
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
	case "tab":
		// Switch between sections
		if m.section == 0 {
			m.section = 1
			m.routingCursor = 0
		} else {
			m.section = 0
			m.cursor = 0
		}
	case "up", "k":
		if m.section == 0 {
			if m.cursor > 0 {
				m.cursor--
			}
		} else {
			if m.routingCursor > 0 {
				m.routingCursor--
			}
		}
	case "down", "j":
		if m.section == 0 {
			if m.cursor < len(m.allConfigs)-1 {
				m.cursor++
			}
		} else {
			if m.routingCursor < len(knownScenarios)-1 {
				m.routingCursor++
			}
		}
	case " ":
		if m.section == 0 {
			// Toggle selection in default providers
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
		}
	case "enter":
		if m.section == 0 {
			// Enter grab mode only if current item is in order
			if m.cursor < len(m.allConfigs) {
				name := m.allConfigs[m.cursor]
				if m.orderIndex(name) > 0 {
					m.grabbed = true
				}
			}
		} else {
			// Toggle scenario expansion or enter scenario editor
			if m.routingCursor < len(knownScenarios) {
				scenario := knownScenarios[m.routingCursor].scenario
				// Enter scenario editor
				return m, func() tea.Msg {
					return switchToScenarioEditMsg{
						profile:  m.profile,
						scenario: scenario,
					}
				}
			}
		}
	case "s", "ctrl+s", "cmd+s":
		return m.saveAndExit()
	}
	return m, nil
}

func (m fallbackModel) saveAndExit() (fallbackModel, tea.Cmd) {
	pc := &config.ProfileConfig{
		Providers: m.order,
	}

	// Build routing config
	if len(m.routingOrder) > 0 {
		pc.Routing = make(map[config.Scenario]*config.ScenarioRoute)
		for scenario, providerNames := range m.routingOrder {
			if len(providerNames) == 0 {
				continue
			}
			var providerRoutes []*config.ProviderRoute
			for _, name := range providerNames {
				pr := &config.ProviderRoute{Name: name}
				if models, ok := m.routingModels[scenario]; ok {
					if model, ok := models[name]; ok && model != "" {
						pr.Model = model
					}
				}
				providerRoutes = append(providerRoutes, pr)
			}
			pc.Routing[scenario] = &config.ScenarioRoute{Providers: providerRoutes}
		}
	}

	if err := config.SetProfileConfig(m.profile, pc); err != nil {
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

// RunEditProfile is the standalone entry point for editing a profile.
func RunEditProfile(profile string) error {
	fm := newFallbackModel(profile)
	wrapper := &fallbackWrapper{fallback: fm}
	p := tea.NewProgram(wrapper, tea.WithAltScreen())
	_, err := p.Run()
	return err
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
		// Default Providers Section
		sectionStyle := sectionTitleStyle
		if m.section != 0 {
			sectionStyle = dimStyle
		}
		content.WriteString(sectionStyle.Render(" Default Providers"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Space to toggle, Enter to reorder"))
		content.WriteString("\n\n")

		for i, name := range m.allConfigs {
			cursor := "  "
			style := tableRowStyle
			if m.section == 0 && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			orderIdx := m.orderIndex(name)
			var checkbox string
			if orderIdx > 0 {
				checkbox = lipgloss.NewStyle().
					Foreground(successColor).
					Render(fmt.Sprintf("[%d]", orderIdx))
			} else {
				checkbox = dimStyle.Render("[ ]")
			}

			grabIndicator := ""
			if m.grabbed && m.section == 0 && i == m.cursor {
				grabIndicator = " " + lipgloss.NewStyle().
					Foreground(accentColor).
					Render("(reordering)")
			}

			line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, grabIndicator)
			content.WriteString(style.Render(line))
			if i < len(m.allConfigs)-1 {
				content.WriteString("\n")
			}
		}

		// Routing Section
		content.WriteString("\n\n")
		sectionStyle = sectionTitleStyle
		if m.section != 1 {
			sectionStyle = dimStyle
		}
		content.WriteString(sectionStyle.Render(" Scenario Routing"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Enter to configure scenario"))
		content.WriteString("\n\n")

		for i, ks := range knownScenarios {
			cursor := "  "
			style := tableRowStyle
			if m.section == 1 && i == m.routingCursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			// Check if configured
			providerCount := 0
			if order, ok := m.routingOrder[ks.scenario]; ok && len(order) > 0 {
				providerCount = len(order)
			}

			// Show provider count if configured
			countInfo := ""
			if providerCount > 0 {
				countInfo = dimStyle.Render(fmt.Sprintf(" (%d providers)", providerCount))
			}

			line := fmt.Sprintf("%s%s%s", cursor, ks.label, countInfo)
			content.WriteString(style.Render(line))
			if i < len(knownScenarios)-1 {
				content.WriteString("\n")
			}
		}
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(50).
		Render(content.String())

	b.WriteString(contentBox)
	b.WriteString("\n\n")

	if m.saved {
		b.WriteString(successStyle.Render("  ✓ " + m.status))
	} else {
		if m.status != "" {
			b.WriteString(errorStyle.Render("  ✗ " + m.status))
			b.WriteString("\n")
		}
		b.WriteString(helpStyle.Render("  tab switch section • s save • esc back"))
	}

	return b.String()
}
