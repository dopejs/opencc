package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

type listViewMode int

const (
	listViewAll listViewMode = iota
	listViewProviders
	listViewGroups
)

type detailListModel struct {
	mode          listViewMode
	providers     []providerDetail
	groups        []groupDetail
	cursor        int
	inProviders   bool // true = cursor in providers section, false = in groups
	showingDetail bool
	detailContent string
}

type providerDetail struct {
	name   string
	config *config.ProviderConfig
	fbIdx  int // index in default group, 0 = not in default
}

type groupDetail struct {
	name  string
	order []string
}

func newDetailListModel(mode listViewMode) detailListModel {
	return detailListModel{
		mode:        mode,
		inProviders: true,
	}
}

type detailListLoadedMsg struct {
	providers []providerDetail
	groups    []groupDetail
}

func (m detailListModel) Init() tea.Cmd {
	mode := m.mode
	return func() tea.Msg {
		var providers []providerDetail
		var groups []groupDetail

		if mode == listViewAll || mode == listViewProviders {
			store := config.DefaultStore()
			names := store.ProviderNames()
			fbOrder, _ := config.ReadFallbackOrder()
			fbMap := make(map[string]int)
			for i, n := range fbOrder {
				fbMap[n] = i + 1
			}

			for _, name := range names {
				p := store.GetProvider(name)
				providers = append(providers, providerDetail{
					name:   name,
					config: p,
					fbIdx:  fbMap[name],
				})
			}
		}

		if mode == listViewAll || mode == listViewGroups {
			names := config.ListProfiles()
			for _, name := range names {
				order, _ := config.ReadProfileOrder(name)
				groups = append(groups, groupDetail{
					name:  name,
					order: order,
				})
			}
		}

		return detailListLoadedMsg{providers: providers, groups: groups}
	}
}

func (m detailListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case detailListLoadedMsg:
		m.providers = msg.providers
		m.groups = msg.groups
		m.cursor = 0
		m.inProviders = len(m.providers) > 0
		return m, nil

	case tea.KeyMsg:
		if m.showingDetail {
			return m.handleDetailKey(msg)
		}
		return m.handleListKey(msg)
	}
	return m, nil
}

func (m detailListModel) handleListKey(msg tea.KeyMsg) (detailListModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		return m, tea.Quit
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "enter":
		m.showingDetail = true
		m.detailContent = m.buildDetail()
	}
	return m, nil
}

func (m detailListModel) handleDetailKey(msg tea.KeyMsg) (detailListModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "enter":
		m.showingDetail = false
	}
	return m, nil
}

func (m *detailListModel) moveCursor(delta int) {
	totalProviders := len(m.providers)
	totalGroups := len(m.groups)

	if m.inProviders {
		newCursor := m.cursor + delta
		if newCursor < 0 {
			return
		}
		if newCursor >= totalProviders {
			if totalGroups > 0 {
				m.inProviders = false
				m.cursor = 0
			}
			return
		}
		m.cursor = newCursor
	} else {
		newCursor := m.cursor + delta
		if newCursor < 0 {
			if totalProviders > 0 {
				m.inProviders = true
				m.cursor = totalProviders - 1
			}
			return
		}
		if newCursor >= totalGroups {
			return
		}
		m.cursor = newCursor
	}
}

func (m detailListModel) buildDetail() string {
	var b strings.Builder

	if m.inProviders && m.cursor < len(m.providers) {
		p := m.providers[m.cursor]
		b.WriteString(fmt.Sprintf("Provider: %s\n", p.name))
		b.WriteString(strings.Repeat("─", 40))
		b.WriteString("\n")
		if p.config != nil {
			b.WriteString(fmt.Sprintf("Base URL:        %s\n", p.config.BaseURL))
			b.WriteString(fmt.Sprintf("Auth Token:      %s\n", p.config.AuthToken))
			b.WriteString(fmt.Sprintf("Model:           %s\n", valueOrDash(p.config.Model)))
			b.WriteString(fmt.Sprintf("Reasoning Model: %s\n", valueOrDash(p.config.ReasoningModel)))
			b.WriteString(fmt.Sprintf("Haiku Model:     %s\n", valueOrDash(p.config.HaikuModel)))
			b.WriteString(fmt.Sprintf("Opus Model:      %s\n", valueOrDash(p.config.OpusModel)))
			b.WriteString(fmt.Sprintf("Sonnet Model:    %s\n", valueOrDash(p.config.SonnetModel)))
		}
		if p.fbIdx > 0 {
			b.WriteString(fmt.Sprintf("\nDefault Group:   #%d\n", p.fbIdx))
		}
	} else if !m.inProviders && m.cursor < len(m.groups) {
		g := m.groups[m.cursor]
		b.WriteString(fmt.Sprintf("Group: %s\n", g.name))
		b.WriteString(strings.Repeat("─", 40))
		b.WriteString("\n")
		if len(g.order) == 0 {
			b.WriteString("No providers in this group.\n")
		} else {
			b.WriteString("Providers:\n")
			for i, name := range g.order {
				b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, name))
			}
		}
	}

	return b.String()
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func (m detailListModel) View() string {
	if m.showingDetail {
		return m.viewDetail()
	}
	return m.viewList()
}

func (m detailListModel) viewList() string {
	width := 80  // default width
	height := 24 // default height

	sidePadding := 2
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("📋 opencc list")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Providers section
	if m.mode == listViewAll || m.mode == listViewProviders {
		providerContent := m.renderProviderList()
		providerBox := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Render(providerContent)
		b.WriteString(providerBox)
		b.WriteString("\n")
	}

	// Groups section
	if m.mode == listViewAll || m.mode == listViewGroups {
		if m.mode == listViewAll {
			b.WriteString("\n")
		}
		groupContent := m.renderGroupList()
		groupBox := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Render(groupContent)
		b.WriteString(groupBox)
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
	helpBar := RenderHelpBar("↑↓ navigate • Enter detail • q quit", width)
	view.WriteString(helpBar)

	return view.String()
}

func (m detailListModel) renderProviderList() string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render(" Providers"))
	b.WriteString("\n")

	if len(m.providers) == 0 {
		b.WriteString(dimStyle.Render("  (none)"))
	} else {
		header := fmt.Sprintf("  %-12s %-4s %-20s %s", "NAME", "GRP", "MODEL", "BASE URL")
		b.WriteString(dimStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + repeatString("─", 60)))
		b.WriteString("\n")

		for i, p := range m.providers {
			cursor := "  "
			style := tableRowStyle
			if m.inProviders && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			grpTag := "  - "
			if p.fbIdx > 0 {
				grpTag = fmt.Sprintf("[%d] ", p.fbIdx)
			}

			model := "-"
			baseURL := "-"
			if p.config != nil {
				if p.config.Model != "" {
					model = p.config.Model
					if len(model) > 18 {
						model = model[:16] + ".."
					}
				}
				if p.config.BaseURL != "" {
					baseURL = p.config.BaseURL
					if len(baseURL) > 25 {
						baseURL = baseURL[:23] + ".."
					}
				}
			}

			line := fmt.Sprintf("%s%-12s %s %-20s %s", cursor, p.name, grpTag, model, baseURL)
			b.WriteString(style.Render(line))
			if i < len(m.providers)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func (m detailListModel) renderGroupList() string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render(" Groups"))
	b.WriteString("\n")

	if len(m.groups) == 0 {
		b.WriteString(dimStyle.Render("  (none)"))
	} else {
		header := fmt.Sprintf("  %-14s %-30s", "NAME", "PROVIDERS")
		b.WriteString(dimStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + repeatString("─", 46)))
		b.WriteString("\n")

		for i, g := range m.groups {
			cursor := "  "
			style := tableRowStyle
			if !m.inProviders && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}

			providers := fmt.Sprintf("%d provider(s)", len(g.order))
			if len(g.order) > 0 && len(g.order) <= 3 {
				providers = strings.Join(g.order, ", ")
			}
			if len(providers) > 28 {
				providers = providers[:26] + ".."
			}

			line := fmt.Sprintf("%s%-14s %-30s", cursor, g.name, providers)
			b.WriteString(style.Render(line))
			if i < len(m.groups)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func (m detailListModel) viewDetail() string {
	width := 80  // default width
	height := 24 // default height

	sidePadding := 2
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("🔍 Detail")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Detail box
	detailBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Render(m.detailContent)
	b.WriteString(detailBox)

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
	helpBar := RenderHelpBar("Esc/Enter back • q quit", width)
	view.WriteString(helpBar)

	return view.String()
}

// RunDetailList runs the detail list TUI.
func RunDetailList(mode listViewMode) error {
	m := newDetailListModel(mode)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// Exported constants for cmd package
const (
	ListViewAll       = listViewAll
	ListViewProviders = listViewProviders
	ListViewGroups    = listViewGroups
)
