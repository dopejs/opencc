package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

// configMainModel is the main config TUI showing providers and groups.
type configMainModel struct {
	providers   []providerEntry
	groups      []groupEntry
	cursor      int
	inProviders bool   // true = cursor in providers section
	cancelled   bool
	deleting    bool   // true = showing delete confirmation
	status      string // status message
}

type providerEntry struct {
	name   string
	config *config.ProviderConfig
}

type groupEntry struct {
	name  string
	count int
}

func newConfigMainModel() configMainModel {
	return configMainModel{inProviders: true}
}

type configMainLoadedMsg struct {
	providers []providerEntry
	groups    []groupEntry
}

func (m configMainModel) Init() tea.Cmd {
	return func() tea.Msg {
		store := config.DefaultStore()
		names := store.ProviderNames()
		var providers []providerEntry
		for _, name := range names {
			providers = append(providers, providerEntry{
				name:   name,
				config: store.GetProvider(name),
			})
		}

		groupNames := config.ListProfiles()
		var groups []groupEntry
		for _, name := range groupNames {
			order, _ := config.ReadProfileOrder(name)
			groups = append(groups, groupEntry{name: name, count: len(order)})
		}

		return configMainLoadedMsg{providers: providers, groups: groups}
	}
}

func (m configMainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case configMainLoadedMsg:
		m.providers = msg.providers
		m.groups = msg.groups
		m.cursor = 0
		m.inProviders = len(m.providers) > 0
		m.deleting = false
		m.status = ""
		return m, nil

	case tea.KeyMsg:
		if m.deleting {
			return m.handleDeleteConfirm(msg)
		}
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			m.moveCursor(-1)
			m.status = ""
		case "down", "j":
			m.moveCursor(1)
			m.status = ""
		case "enter", "e":
			return m, m.selectCurrent()
		case "a":
			return m, func() tea.Msg { return configAddMsg{} }
		case "d":
			return m.startDelete()
		}
	case configReturnMsg:
		return m, m.Init()
	}
	return m, nil
}

func (m configMainModel) startDelete() (configMainModel, tea.Cmd) {
	if m.inProviders {
		if len(m.providers) <= 1 {
			m.status = "Cannot delete the last provider"
			return m, nil
		}
		m.deleting = true
	} else {
		if m.cursor < len(m.groups) && m.groups[m.cursor].name == "default" {
			m.status = "Cannot delete the default group"
			return m, nil
		}
		m.deleting = true
	}
	return m, nil
}

func (m configMainModel) handleDeleteConfirm(msg tea.KeyMsg) (configMainModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.inProviders && m.cursor < len(m.providers) {
			name := m.providers[m.cursor].name
			config.DeleteProviderByName(name)
			m.deleting = false
			return m, m.Init()
		}
		if !m.inProviders && m.cursor < len(m.groups) {
			name := m.groups[m.cursor].name
			config.DeleteProfile(name)
			m.deleting = false
			return m, m.Init()
		}
	case "n", "N", "esc":
		m.deleting = false
	}
	return m, nil
}

type configAddMsg struct{}
type configReturnMsg struct{}

func (m *configMainModel) moveCursor(delta int) {
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

func (m configMainModel) selectCurrent() tea.Cmd {
	if m.inProviders && m.cursor < len(m.providers) {
		name := m.providers[m.cursor].name
		return func() tea.Msg { return configEditProviderMsg{name: name} }
	}
	if !m.inProviders && m.cursor < len(m.groups) {
		name := m.groups[m.cursor].name
		return func() tea.Msg { return configEditGroupMsg{name: name} }
	}
	return nil
}

type configEditProviderMsg struct{ name string }
type configEditGroupMsg struct{ name string }

func (m configMainModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("⚙  opencc config")
	b.WriteString(header)
	b.WriteString("\n\n")

	// Providers section
	providerContent := m.renderProviders()
	providerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(50).
		Render(providerContent)
	b.WriteString(providerBox)
	b.WriteString("\n\n")

	// Groups section
	groupContent := m.renderGroups()
	groupBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(50).
		Render(groupContent)
	b.WriteString(groupBox)
	b.WriteString("\n\n")

	// Status/Delete confirmation
	if m.deleting {
		var name string
		if m.inProviders && m.cursor < len(m.providers) {
			name = m.providers[m.cursor].name
		} else if !m.inProviders && m.cursor < len(m.groups) {
			name = m.groups[m.cursor].name
		}
		confirmBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(0, 1).
			Render(errorStyle.Render(fmt.Sprintf(" Delete '%s'? (y/n) ", name)))
		b.WriteString(confirmBox)
	} else {
		if m.status != "" {
			b.WriteString(errorStyle.Render("  " + m.status))
			b.WriteString("\n")
		}
		help := helpStyle.Render("  ↑↓ navigate • enter edit • a add • d delete • q quit")
		b.WriteString(help)
	}

	return b.String()
}

func (m configMainModel) renderProviders() string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render(" Providers"))
	b.WriteString("\n")

	if len(m.providers) == 0 {
		b.WriteString(dimStyle.Render("  (none)"))
	} else {
		// Header
		header := fmt.Sprintf("  %-14s %-24s", "NAME", "MODEL")
		b.WriteString(dimStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + repeatString("─", 40)))
		b.WriteString("\n")

		for i, p := range m.providers {
			cursor := "  "
			style := tableRowStyle
			if m.inProviders && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}
			model := "-"
			if p.config != nil && p.config.Model != "" {
				model = p.config.Model
				if len(model) > 22 {
					model = model[:20] + ".."
				}
			}
			line := fmt.Sprintf("%s%-14s %-24s", cursor, p.name, model)
			b.WriteString(style.Render(line))
			if i < len(m.providers)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func (m configMainModel) renderGroups() string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render(" Groups"))
	b.WriteString("\n")

	if len(m.groups) == 0 {
		b.WriteString(dimStyle.Render("  (none)"))
	} else {
		// Header
		header := fmt.Sprintf("  %-14s %-20s", "NAME", "PROVIDERS")
		b.WriteString(dimStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + repeatString("─", 36)))
		b.WriteString("\n")

		for i, g := range m.groups {
			cursor := "  "
			style := tableRowStyle
			if !m.inProviders && i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}
			count := fmt.Sprintf("%d provider(s)", g.count)
			line := fmt.Sprintf("%s%-14s %-20s", cursor, g.name, count)
			b.WriteString(style.Render(line))
			if i < len(m.groups)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

// configMainWrapper wraps configMainModel to handle sub-editors.
type configMainWrapper struct {
	main       configMainModel
	subEditor  tea.Model
	inSubEdit  bool
	cancelled  bool
}

func newConfigMainWrapper() configMainWrapper {
	return configMainWrapper{
		main: newConfigMainModel(),
	}
}

func (m configMainWrapper) Init() tea.Cmd {
	return m.main.Init()
}

func (m configMainWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inSubEdit {
		switch msg.(type) {
		case switchToListMsg:
			// Sub-editor finished
			m.inSubEdit = false
			m.subEditor = nil
			return m, m.main.Init()
		case configEditProviderMsg, configEditGroupMsg, configAddGroupMsg:
			// Transition from addTypeSelector to actual editor
			m.inSubEdit = false
		case switchToRoutingMsg:
			// Routing editor requested from within fallback editor
			m.inSubEdit = false
		default:
			var cmd tea.Cmd
			m.subEditor, cmd = m.subEditor.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	case configEditProviderMsg:
		m.inSubEdit = true
		if msg.name == "" {
			// New provider — use wrapper that chains profile picker
			m.subEditor = newEditorWithProfilePickerWrapper()
			return m, m.subEditor.Init()
		}
		editor := newEditorModel(msg.name)
		m.subEditor = &editorWrapper{editor: editor}
		return m, editor.init()
	case configEditGroupMsg:
		m.inSubEdit = true
		fb := newFallbackModel(msg.name)
		m.subEditor = &fallbackWrapper{fallback: fb}
		return m, fb.init()
	case configAddMsg:
		m.inSubEdit = true
		m.subEditor = newAddTypeSelector()
		return m, nil
	case configAddGroupMsg:
		m.inSubEdit = true
		gc := newGroupCreateModel("")
		m.subEditor = &groupCreateWrapper{model: gc}
		return m, gc.Init()
	case switchToRoutingMsg:
		m.inSubEdit = true
		rm := newRoutingModel(msg.profile)
		m.subEditor = &routingWrapper{routing: rm}
		return m, rm.init()
	}

	var cmd tea.Cmd
	newMain, cmd := m.main.Update(msg)
	m.main = newMain.(configMainModel)
	return m, cmd
}

func (m configMainWrapper) View() string {
	if m.inSubEdit && m.subEditor != nil {
		return m.subEditor.View()
	}
	return m.main.View()
}

// editorWrapper wraps editorModel for use in configMainWrapper.
type editorWrapper struct {
	editor editorModel
}

func (w *editorWrapper) Init() tea.Cmd {
	return w.editor.init()
}

func (w *editorWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.editor, cmd = w.editor.update(msg)
	return w, cmd
}

func (w *editorWrapper) View() string {
	return w.editor.view(0, 0)
}

// editorWithProfilePickerWrapper wraps editorModel and chains a profile
// multi-select after saving a new provider.
type editorWithProfilePickerWrapper struct {
	editor       editorModel
	profilePick  profileMultiSelectModel
	phase        int // 0=editor, 1=profile picker
	providerName string
}

func newEditorWithProfilePickerWrapper() *editorWithProfilePickerWrapper {
	return &editorWithProfilePickerWrapper{
		editor: newEditorModel(""),
	}
}

func (w *editorWithProfilePickerWrapper) Init() tea.Cmd {
	return w.editor.init()
}

func (w *editorWithProfilePickerWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if w.phase == 0 {
		switch msg.(type) {
		case switchToListMsg:
			// Editor finished — check if we created a new provider and have profiles
			w.providerName = w.editor.createdName
			profiles := config.ListProfiles()
			if w.providerName != "" && len(profiles) > 0 {
				w.phase = 1
				w.profilePick = newProfileMultiSelectModel()
				return w, nil
			}
			// No profiles or was editing — go back to list
			return w, func() tea.Msg { return switchToListMsg{} }
		}
		var cmd tea.Cmd
		w.editor, cmd = w.editor.update(msg)
		return w, cmd
	}

	// Phase 1: profile picker
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return w, func() tea.Msg { return switchToListMsg{} }
		case "esc", "q":
			// Skip — don't add to any profile
			return w, func() tea.Msg { return switchToListMsg{} }
		case "up", "k":
			if w.profilePick.cursor > 0 {
				w.profilePick.cursor--
			}
		case "down", "j":
			if w.profilePick.cursor < len(w.profilePick.profiles)-1 {
				w.profilePick.cursor++
			}
		case " ":
			if w.profilePick.cursor < len(w.profilePick.profiles) {
				name := w.profilePick.profiles[w.profilePick.cursor]
				if w.profilePick.selected[name] {
					delete(w.profilePick.selected, name)
				} else {
					w.profilePick.selected[name] = true
				}
			}
		case "enter":
			// Confirm — add provider to selected profiles
			for _, profile := range w.profilePick.profiles {
				if w.profilePick.selected[profile] {
					order, _ := config.ReadProfileOrder(profile)
					order = append(order, w.providerName)
					config.WriteProfileOrder(profile, order)
				}
			}
			return w, func() tea.Msg { return switchToListMsg{} }
		}
	}
	return w, nil
}

func (w *editorWithProfilePickerWrapper) View() string {
	if w.phase == 0 {
		return w.editor.view(0, 0)
	}
	return w.profilePick.View()
}

// fallbackWrapper wraps fallbackModel for use in configMainWrapper.
type fallbackWrapper struct {
	fallback fallbackModel
}

func (w *fallbackWrapper) Init() tea.Cmd {
	return w.fallback.init()
}

func (w *fallbackWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.fallback, cmd = w.fallback.update(msg)
	return w, cmd
}

func (w *fallbackWrapper) View() string {
	return w.fallback.view(0, 0)
}

// groupCreateWrapper wraps groupCreateModel for inline use.
// After successful creation, transitions to the fallback editor.
type groupCreateWrapper struct {
	model    groupCreateModel
	fallback *fallbackWrapper
	phase    int // 0=name input, 1=fallback editor
}

func (w *groupCreateWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *groupCreateWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if w.phase == 1 {
		result, cmd := w.fallback.Update(msg)
		w.fallback = result.(*fallbackWrapper)
		return w, cmd
	}

	result, cmd := w.model.Update(msg)
	w.model = result.(groupCreateModel)
	if w.model.cancelled {
		return w, func() tea.Msg { return switchToListMsg{} }
	}
	if w.model.created != "" {
		// Switch to fallback editor for the new group
		w.phase = 1
		fb := newFallbackModel(w.model.created)
		w.fallback = &fallbackWrapper{fallback: fb}
		return w, fb.init()
	}
	return w, cmd
}

func (w *groupCreateWrapper) View() string {
	if w.phase == 1 {
		return w.fallback.View()
	}
	return w.model.View()
}

// addTypeSelector lets user choose provider or group to add.
type addTypeSelector struct {
	items    []string
	cursor   int
	selected string
}

func newAddTypeSelector() *addTypeSelector {
	return &addTypeSelector{
		items: []string{"provider", "group"},
	}
}

func (s *addTypeSelector) Init() tea.Cmd {
	return nil
}

func (s *addTypeSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, func() tea.Msg { return switchToListMsg{} }
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.items)-1 {
				s.cursor++
			}
		case "enter":
			s.selected = s.items[s.cursor]
			if s.selected == "provider" {
				return s, func() tea.Msg { return configEditProviderMsg{name: ""} }
			}
			return s, func() tea.Msg { return configAddGroupMsg{} }
		}
	}
	return s, nil
}

type configAddGroupMsg struct{}

func (s *addTypeSelector) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  Add"))
	b.WriteString("\n\n")
	for i, item := range s.items {
		cursor := "  "
		style := dimStyle
		if i == s.cursor {
			cursor = "▸ "
			style = selectedStyle
		}
		b.WriteString(style.Render(cursor + item))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  enter:select  esc:back"))
	return b.String()
}

// RunConfigMain runs the main config TUI.
func RunConfigMain() error {
	m := newConfigMainWrapper()
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}
	if wrapper, ok := result.(configMainWrapper); ok && wrapper.cancelled {
		return fmt.Errorf("cancelled")
	}
	return nil
}
