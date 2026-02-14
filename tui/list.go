package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dopejs/opencc/internal/config"
)

type listItem struct {
	Name   string
	Config *config.ProviderConfig
}

type listModel struct {
	configs  []listItem
	fbOrder  map[string]int
	cursor   int
	status   string
	deleting bool // confirm delete mode
}

func newListModel() listModel {
	return listModel{}
}

type configsLoadedMsg struct {
	configs []listItem
	fbOrder map[string]int
}

func (m listModel) init() tea.Cmd {
	return func() tea.Msg {
		store := config.DefaultStore()
		providerMap := store.ProviderMap()
		var items []listItem
		for name, p := range providerMap {
			items = append(items, listItem{Name: name, Config: p})
		}
		fbNames, _ := config.ReadFallbackOrder()
		fbOrder := make(map[string]int)
		for i, n := range fbNames {
			fbOrder[n] = i + 1
		}
		return configsLoadedMsg{configs: items, fbOrder: fbOrder}
	}
}

func (m listModel) update(msg tea.Msg) (listModel, tea.Cmd) {
	switch msg := msg.(type) {
	case configsLoadedMsg:
		m.configs = msg.configs
		m.fbOrder = msg.fbOrder
		// Sort: fallback configs first (by order), then the rest alphabetically
		sort.Slice(m.configs, func(i, j int) bool {
			fi, oki := m.fbOrder[m.configs[i].Name]
			fj, okj := m.fbOrder[m.configs[j].Name]
			if oki && okj {
				return fi < fj
			}
			if oki {
				return true
			}
			if okj {
				return false
			}
			return m.configs[i].Name < m.configs[j].Name
		})
		m.cursor = 0
		m.deleting = false
		return m, nil

	case statusMsg:
		m.status = msg.text
		return m, nil

	case tea.KeyMsg:
		if m.deleting {
			return m.handleDeleteConfirm(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m listModel) handleKey(msg tea.KeyMsg) (listModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.configs)-1 {
			m.cursor++
		}
	case "a":
		// Add new config
		return m, func() tea.Msg { return switchToEditorMsg{} }
	case "e", "enter":
		if len(m.configs) > 0 {
			name := m.configs[m.cursor].Name
			return m, func() tea.Msg { return switchToEditorMsg{configName: name} }
		}
	case "d":
		if len(m.configs) > 0 {
			m.deleting = true
		}
	case "f":
		return m, func() tea.Msg { return switchToProfileListMsg{} }
	}
	return m, nil
}

func (m listModel) handleDeleteConfirm(msg tea.KeyMsg) (listModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.configs) {
			name := m.configs[m.cursor].Name
			config.DeleteProviderByName(name)
			m.deleting = false
			return m, m.init()
		}
	case "n", "N", "esc":
		m.deleting = false
	}
	return m, nil
}

func (m listModel) view(width, height int) string {
	sidePadding := 2
	var b strings.Builder

	b.WriteString(titleStyle.Render("  opencc configurations"))
	b.WriteString("\n\n")

	if len(m.configs) == 0 {
		b.WriteString("  No configurations found.\n")
		b.WriteString("  Press 'a' to add a new configuration.\n")
	} else {
		for i, item := range m.configs {
			cursor := "  "
			style := dimStyle
			if i == m.cursor {
				cursor = "▸ "
				style = selectedStyle
			}

			baseURL := item.Config.BaseURL
			model := item.Config.Model
			if model == "" {
				model = "-"
			}

			fbTag := ""
			if idx, ok := m.fbOrder[item.Name]; ok {
				fbTag = fmt.Sprintf(" [fb:%d]", idx)
			}

			line := fmt.Sprintf("%s%-12s model=%-20s  %s%s", cursor, item.Name, model, baseURL, fbTag)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	if m.deleting && m.cursor < len(m.configs) {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  Delete '%s'? (y/n)", m.configs[m.cursor].Name)))
	}

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("  " + m.status))
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
	helpBar := RenderHelpBar("a add • e/Enter edit • d delete • f fallback profiles • q quit", width)
	view.WriteString(helpBar)

	return view.String()
}
