package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

// switchToRoutingMsg triggers opening the routing editor from the fallback editor.
type switchToRoutingMsg struct {
	profile string
}

// switchToScenarioEditMsg triggers opening a specific scenario editor from fallback.
type switchToScenarioEditMsg struct {
	profile  string
	scenario config.Scenario
}

// scenarioEntry represents one scenario row in the routing editor.
type scenarioEntry struct {
	scenario   config.Scenario
	label      string
	configured bool // has an existing route
}

// routingModel is the TUI for editing scenario routing rules.
type routingModel struct {
	profile      string
	scenarios    []scenarioEntry
	cursor       int
	phase        int // 0=scenario list, 1=edit scenario
	editModel    scenarioEditModel
	allProviders []string
	saved        bool
	status       string
}

// scenarioEditModel edits a single scenario's providers and per-provider models.
type scenarioEditModel struct {
	scenario        config.Scenario
	allProviders    []string
	order           []string          // selected providers for this scenario
	providerModels  map[string]string // provider name → model override
	cursor          int
	phase           int    // 0=select providers, 1=edit provider model
	editingProvider string // provider being edited in phase 1
	modelInput      string
	modelCursor     int
}

var knownScenarios = []struct {
	scenario config.Scenario
	label    string
}{
	{config.ScenarioThink, "think      (thinking mode requests)"},
	{config.ScenarioImage, "image      (requests with images)"},
	{config.ScenarioLongContext, "longContext (>32k chars total)"},
}

func newRoutingModel(profile string) routingModel {
	return routingModel{
		profile: profile,
	}
}

type routingLoadedMsg struct {
	scenarios    []scenarioEntry
	allProviders []string
	routing      map[config.Scenario]*config.ScenarioRoute
}

func (m routingModel) init() tea.Cmd {
	profile := m.profile
	return func() tea.Msg {
		pc := config.GetProfileConfig(profile)
		allProviders := config.ProviderNames()

		var routing map[config.Scenario]*config.ScenarioRoute
		if pc != nil {
			routing = pc.Routing
		}

		var scenarios []scenarioEntry
		for _, ks := range knownScenarios {
			configured := false
			if routing != nil {
				if _, ok := routing[ks.scenario]; ok {
					configured = true
				}
			}
			scenarios = append(scenarios, scenarioEntry{
				scenario:   ks.scenario,
				label:      ks.label,
				configured: configured,
			})
		}

		return routingLoadedMsg{
			scenarios:    scenarios,
			allProviders: allProviders,
			routing:      routing,
		}
	}
}

func (m routingModel) update(msg tea.Msg) (routingModel, tea.Cmd) {
	if m.saved {
		if _, ok := msg.(saveExitMsg); ok {
			return m, func() tea.Msg { return switchToListMsg{} }
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case routingLoadedMsg:
		m.scenarios = msg.scenarios
		m.allProviders = msg.allProviders
		m.cursor = 0
		m.phase = 0
		return m, nil

	case tea.KeyMsg:
		if m.phase == 1 {
			return m.updateEditScenario(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m routingModel) handleKey(msg tea.KeyMsg) (routingModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return m, func() tea.Msg { return switchToListMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.scenarios)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(m.scenarios) {
			s := m.scenarios[m.cursor]
			m.phase = 1
			m.editModel = newScenarioEditModel(s.scenario, m.allProviders, m.profile)
		}
	case "x":
		// Clear route for current scenario
		if m.cursor < len(m.scenarios) {
			s := m.scenarios[m.cursor]
			pc := config.GetProfileConfig(m.profile)
			if pc != nil && pc.Routing != nil {
				delete(pc.Routing, s.scenario)
				if len(pc.Routing) == 0 {
					pc.Routing = nil
				}
				config.SetProfileConfig(m.profile, pc)
				m.scenarios[m.cursor].configured = false
				m.status = fmt.Sprintf("Cleared %s route", s.scenario)
			}
		}
	case "s", "ctrl+s", "cmd+s":
		m.saved = true
		m.status = "Saved"
		return m, saveExitTick()
	}
	return m, nil
}

func (m routingModel) updateEditScenario(msg tea.KeyMsg) (routingModel, tea.Cmd) {
	em := &m.editModel

	if em.phase == 1 {
		// Model editing phase for a specific provider
		switch msg.String() {
		case "esc":
			em.phase = 0
			em.editingProvider = ""
			em.modelInput = ""
		case "enter":
			// Save model for this provider
			if em.editingProvider != "" {
				if em.providerModels == nil {
					em.providerModels = make(map[string]string)
				}
				trimmed := strings.TrimSpace(em.modelInput)
				if trimmed != "" {
					em.providerModels[em.editingProvider] = trimmed
				} else {
					delete(em.providerModels, em.editingProvider)
				}
			}
			em.phase = 0
			em.editingProvider = ""
			em.modelInput = ""
		case "backspace":
			if len(em.modelInput) > 0 {
				em.modelInput = em.modelInput[:len(em.modelInput)-1]
			}
		default:
			if len(msg.String()) == 1 {
				em.modelInput += msg.String()
			}
		}
		return m, nil
	}

	// Phase 0: provider selection
	switch msg.String() {
	case "esc":
		m.phase = 0
	case "up", "k":
		if em.cursor > 0 {
			em.cursor--
		}
	case "down", "j":
		if em.cursor < len(em.allProviders)-1 {
			em.cursor++
		}
	case " ":
		// Toggle provider selection
		if em.cursor < len(em.allProviders) {
			name := em.allProviders[em.cursor]
			if idx := scenarioOrderIndex(em.order, name); idx >= 0 {
				em.order = removeFromScenarioOrder(em.order, name)
				// Also remove model override if any
				if em.providerModels != nil {
					delete(em.providerModels, name)
				}
			} else {
				em.order = append(em.order, name)
			}
		}
	case "m":
		// Edit model for selected provider
		if em.cursor < len(em.allProviders) {
			name := em.allProviders[em.cursor]
			// Only allow editing if provider is selected
			if scenarioOrderIndex(em.order, name) >= 0 {
				em.phase = 1
				em.editingProvider = name
				if em.providerModels != nil {
					em.modelInput = em.providerModels[name]
				} else {
					em.modelInput = ""
				}
			}
		}
	case "enter":
		// Save this scenario route
		m.saveScenarioRoute()
		m.phase = 0
		return m, m.init()
	}
	return m, nil
}

func (m *routingModel) saveScenarioRoute() {
	em := m.editModel
	pc := config.GetProfileConfig(m.profile)
	if pc == nil {
		pc = &config.ProfileConfig{Providers: []string{}}
	}
	if pc.Routing == nil {
		pc.Routing = make(map[config.Scenario]*config.ScenarioRoute)
	}

	if len(em.order) == 0 {
		// No providers selected — remove the route
		delete(pc.Routing, em.scenario)
		if len(pc.Routing) == 0 {
			pc.Routing = nil
		}
	} else {
		var providerRoutes []*config.ProviderRoute
		for _, name := range em.order {
			pr := &config.ProviderRoute{Name: name}
			if em.providerModels != nil {
				if model, ok := em.providerModels[name]; ok && model != "" {
					pr.Model = model
				}
			}
			providerRoutes = append(providerRoutes, pr)
		}
		pc.Routing[em.scenario] = &config.ScenarioRoute{
			Providers: providerRoutes,
		}
	}
	config.SetProfileConfig(m.profile, pc)
}

func newScenarioEditModel(scenario config.Scenario, allProviders []string, profile string) scenarioEditModel {
	em := scenarioEditModel{
		scenario:       scenario,
		allProviders:   allProviders,
		providerModels: make(map[string]string),
	}

	// Load existing route data
	pc := config.GetProfileConfig(profile)
	if pc != nil && pc.Routing != nil {
		if route, ok := pc.Routing[scenario]; ok {
			em.order = route.ProviderNames()
			for _, pr := range route.Providers {
				if pr.Model != "" {
					em.providerModels[pr.Name] = pr.Model
				}
			}
		}
	}
	return em
}

func scenarioOrderIndex(order []string, name string) int {
	for i, n := range order {
		if n == name {
			return i
		}
	}
	return -1
}

func removeFromScenarioOrder(order []string, name string) []string {
	var result []string
	for _, n := range order {
		if n != name {
			result = append(result, n)
		}
	}
	return result
}

func (m routingModel) view(width, height int) string {
	var b strings.Builder

	// Header
	title := fmt.Sprintf("Routing: %s", m.profile)
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("🔀 " + title)
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.phase == 1 {
		// Editing a scenario
		b.WriteString(m.renderScenarioEdit())
	} else {
		// Scenario list
		b.WriteString(m.renderScenarioList())
	}

	b.WriteString("\n\n")

	if m.saved {
		b.WriteString(successStyle.Render("  ✓ " + m.status))
	} else {
		if m.status != "" {
			b.WriteString(successStyle.Render("  ✓ " + m.status))
			b.WriteString("\n")
		}
		if m.phase == 0 {
			b.WriteString(helpStyle.Render("  ↑↓ move • enter edit • x clear • s save • esc back"))
		}
	}

	return b.String()
}

func (m routingModel) renderScenarioList() string {
	var content strings.Builder
	content.WriteString(sectionTitleStyle.Render(" Scenario Routes"))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render(" Configure provider chains per request type"))
	content.WriteString("\n\n")

	for i, s := range m.scenarios {
		cursor := "  "
		style := tableRowStyle
		if i == m.cursor {
			cursor = "▸ "
			style = tableSelectedRowStyle
		}

		status := dimStyle.Render("[ ]")
		if s.configured {
			status = lipgloss.NewStyle().
				Foreground(successColor).
				Render("[✓]")
		}

		line := fmt.Sprintf("%s%s %s", cursor, status, s.label)
		content.WriteString(style.Render(line))
		if i < len(m.scenarios)-1 {
			content.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(50).
		Render(content.String())
}

func (m routingModel) renderScenarioEdit() string {
	em := m.editModel
	var content strings.Builder

	scenarioLabel := string(em.scenario)
	content.WriteString(sectionTitleStyle.Render(fmt.Sprintf(" Edit: %s", scenarioLabel)))
	content.WriteString("\n")

	if em.phase == 0 {
		content.WriteString(dimStyle.Render(" Space toggle • m edit model • enter save • esc back"))
		content.WriteString("\n\n")

		// Provider list with per-provider models
		for i, name := range em.allProviders {
			cursor := "  "
			style := tableRowStyle
			if i == em.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			idx := scenarioOrderIndex(em.order, name)
			var checkbox string
			if idx >= 0 {
				checkbox = lipgloss.NewStyle().
					Foreground(successColor).
					Render(fmt.Sprintf("[%d]", idx+1))
			} else {
				checkbox = dimStyle.Render("[ ]")
			}

			// Show model override if configured
			modelInfo := ""
			if idx >= 0 && em.providerModels != nil {
				if model, ok := em.providerModels[name]; ok && model != "" {
					modelInfo = dimStyle.Render(fmt.Sprintf(" (model: %s)", model))
				}
			}

			line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, modelInfo)
			content.WriteString(style.Render(line))
			if i < len(em.allProviders)-1 {
				content.WriteString("\n")
			}
		}
	} else {
		// Model editing phase for specific provider
		content.WriteString(dimStyle.Render(fmt.Sprintf(" Editing model for: %s", em.editingProvider)))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Type model name • enter save • esc cancel"))
		content.WriteString("\n\n")

		content.WriteString(sectionTitleStyle.Render(" Model Override"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Leave empty to use provider's model mapping"))
		content.WriteString("\n\n")

		modelDisplay := em.modelInput + "█"
		content.WriteString(lipgloss.NewStyle().
			Foreground(accentColor).
			Render("  " + modelDisplay))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(60).
		Render(content.String())
}

// routingWrapper wraps routingModel for use in configMainWrapper.
type routingWrapper struct {
	routing routingModel
}

func (w *routingWrapper) Init() tea.Cmd {
	return w.routing.init()
}

func (w *routingWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.routing, cmd = w.routing.update(msg)
	return w, cmd
}

func (w *routingWrapper) View() string {
	return w.routing.view(0, 0)
}

// scenarioEditWrapper wraps scenarioEditModel for standalone use from fallback editor.
type scenarioEditWrapper struct {
	edit    scenarioEditModel
	profile string
}

func (w *scenarioEditWrapper) Init() tea.Cmd {
	return nil
}

func (w *scenarioEditWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if w.edit.phase == 1 {
			// Model editing phase
			switch msg.String() {
			case "esc":
				w.edit.phase = 0
				w.edit.editingProvider = ""
				w.edit.modelInput = ""
			case "enter":
				if w.edit.editingProvider != "" {
					if w.edit.providerModels == nil {
						w.edit.providerModels = make(map[string]string)
					}
					trimmed := strings.TrimSpace(w.edit.modelInput)
					if trimmed != "" {
						w.edit.providerModels[w.edit.editingProvider] = trimmed
					} else {
						delete(w.edit.providerModels, w.edit.editingProvider)
					}
				}
				w.edit.phase = 0
				w.edit.editingProvider = ""
				w.edit.modelInput = ""
			case "backspace":
				if len(w.edit.modelInput) > 0 {
					w.edit.modelInput = w.edit.modelInput[:len(w.edit.modelInput)-1]
				}
			default:
				if len(msg.String()) == 1 {
					w.edit.modelInput += msg.String()
				}
			}
			return w, nil
		}

		// Phase 0: provider selection
		switch msg.String() {
		case "esc":
			// Return to fallback editor without saving
			return w, func() tea.Msg { return switchToFallbackMsg{profile: w.profile} }
		case "enter":
			// Save and return
			pc := config.GetProfileConfig(w.profile)
			if pc == nil {
				pc = &config.ProfileConfig{Providers: []string{}}
			}
			if pc.Routing == nil {
				pc.Routing = make(map[config.Scenario]*config.ScenarioRoute)
			}

			if len(w.edit.order) == 0 {
				delete(pc.Routing, w.edit.scenario)
				if len(pc.Routing) == 0 {
					pc.Routing = nil
				}
			} else {
				var providerRoutes []*config.ProviderRoute
				for _, name := range w.edit.order {
					pr := &config.ProviderRoute{Name: name}
					if w.edit.providerModels != nil {
						if model, ok := w.edit.providerModels[name]; ok && model != "" {
							pr.Model = model
						}
					}
					providerRoutes = append(providerRoutes, pr)
				}
				pc.Routing[w.edit.scenario] = &config.ScenarioRoute{
					Providers: providerRoutes,
				}
			}
			config.SetProfileConfig(w.profile, pc)
			return w, func() tea.Msg { return switchToFallbackMsg{profile: w.profile} }
		case "up", "k":
			if w.edit.cursor > 0 {
				w.edit.cursor--
			}
		case "down", "j":
			if w.edit.cursor < len(w.edit.allProviders)-1 {
				w.edit.cursor++
			}
		case " ":
			if w.edit.cursor < len(w.edit.allProviders) {
				name := w.edit.allProviders[w.edit.cursor]
				if idx := scenarioOrderIndex(w.edit.order, name); idx >= 0 {
					w.edit.order = removeFromScenarioOrder(w.edit.order, name)
					if w.edit.providerModels != nil {
						delete(w.edit.providerModels, name)
					}
				} else {
					w.edit.order = append(w.edit.order, name)
				}
			}
		case "m":
			if w.edit.cursor < len(w.edit.allProviders) {
				name := w.edit.allProviders[w.edit.cursor]
				if scenarioOrderIndex(w.edit.order, name) >= 0 {
					w.edit.phase = 1
					w.edit.editingProvider = name
					if w.edit.providerModels != nil {
						w.edit.modelInput = w.edit.providerModels[name]
					} else {
						w.edit.modelInput = ""
					}
				}
			}
		}
	}

	return w, nil
}

func (w *scenarioEditWrapper) View() string {
	var b strings.Builder

	scenarioLabel := string(w.edit.scenario)
	title := fmt.Sprintf("Edit Scenario: %s", scenarioLabel)
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("🔀 " + title)
	b.WriteString(header)
	b.WriteString("\n\n")

	var content strings.Builder

	if w.edit.phase == 0 {
		content.WriteString(dimStyle.Render(" Space toggle • m edit model • enter save • esc back"))
		content.WriteString("\n\n")

		for i, name := range w.edit.allProviders {
			cursor := "  "
			style := tableRowStyle
			if i == w.edit.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			idx := scenarioOrderIndex(w.edit.order, name)
			var checkbox string
			if idx >= 0 {
				checkbox = lipgloss.NewStyle().
					Foreground(successColor).
					Render(fmt.Sprintf("[%d]", idx+1))
			} else {
				checkbox = dimStyle.Render("[ ]")
			}

			modelInfo := ""
			if idx >= 0 && w.edit.providerModels != nil {
				if model, ok := w.edit.providerModels[name]; ok && model != "" {
					modelInfo = dimStyle.Render(fmt.Sprintf(" (model: %s)", model))
				}
			}

			line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, modelInfo)
			content.WriteString(style.Render(line))
			if i < len(w.edit.allProviders)-1 {
				content.WriteString("\n")
			}
		}
	} else {
		content.WriteString(dimStyle.Render(fmt.Sprintf(" Editing model for: %s", w.edit.editingProvider)))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Type model name • enter save • esc cancel"))
		content.WriteString("\n\n")

		content.WriteString(sectionTitleStyle.Render(" Model Override"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" Leave empty to use provider's model mapping"))
		content.WriteString("\n\n")

		modelDisplay := w.edit.modelInput + "█"
		content.WriteString(lipgloss.NewStyle().
			Foreground(accentColor).
			Render("  " + modelDisplay))
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(60).
		Render(content.String())

	b.WriteString(contentBox)
	return b.String()
}
