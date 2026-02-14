package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/dopejs/opencc/internal/config"
)

type editorField int

const (
	fieldName editorField = iota
	fieldType  // API type: anthropic or openai
	fieldBaseURL
	fieldAuthToken
	fieldModel
	fieldReasoningModel
	fieldHaikuModel
	fieldOpusModel
	fieldSonnetModel
	fieldEnvVars // special field - opens env vars editor
	fieldCount
)

type editorModel struct {
	fields          [fieldCount]textinput.Model
	focus           editorField
	editing         string // config name being edited, empty = new
	initMode        bool   // true = auto-add to default profile on save (first provider)
	err             string
	saved           bool   // true = save succeeded, waiting to exit
	status          string // "Saved" success message
	createdName     string // name of provider after save (for callers)
	claudeEnvVars   map[string]string // Claude Code environment variables
	codexEnvVars    map[string]string // Codex environment variables
	opencodeEnvVars map[string]string // OpenCode environment variables
	currentEnvCLI   int               // 0=claude, 1=codex, 2=opencode
	envVarsEdit     bool              // true = editing env vars
	envVarsModel    envVarsEditorModel
	providerType    int // 0 = anthropic, 1 = openai
}

func newEditorModel(configName string) editorModel {
	return newEditorModelWithPreset(configName, "")
}

func newEditorModelWithPreset(configName string, presetName string) editorModel {
	var fields [fieldCount]textinput.Model

	for i := range fields {
		fields[i] = textinput.New()
		fields[i].CharLimit = 256
	}

	fields[fieldName].Placeholder = "config name (e.g. work)"
	fields[fieldName].Prompt = "  Name:             "
	// fieldType is handled specially (not a textinput)
	fields[fieldType].Placeholder = ""
	fields[fieldType].Prompt = ""
	fields[fieldBaseURL].Placeholder = "https://api.example.com"
	fields[fieldBaseURL].Prompt = "  Base URL:         "
	fields[fieldAuthToken].Placeholder = "sk-..."
	fields[fieldAuthToken].Prompt = "  Auth Token:       "
	fields[fieldModel].Placeholder = "claude-sonnet-4-5"
	fields[fieldModel].Prompt = "  Model:            "
	fields[fieldReasoningModel].Placeholder = "claude-sonnet-4-5"
	fields[fieldReasoningModel].Prompt = "  Reasoning Model:  "
	fields[fieldHaikuModel].Placeholder = "claude-haiku-4-5"
	fields[fieldHaikuModel].Prompt = "  Haiku Model:      "
	fields[fieldOpusModel].Placeholder = "claude-opus-4-5"
	fields[fieldOpusModel].Prompt = "  Opus Model:       "
	fields[fieldSonnetModel].Placeholder = "claude-sonnet-4-5"
	fields[fieldSonnetModel].Prompt = "  Sonnet Model:     "
	// fieldEnvVars is a special field, not a textinput

	m := editorModel{
		fields:          fields,
		editing:         configName,
		claudeEnvVars:   make(map[string]string),
		codexEnvVars:    make(map[string]string),
		opencodeEnvVars: make(map[string]string),
		currentEnvCLI:   0, // default to claude
		providerType:    0, // default to anthropic
	}

	if configName != "" {
		// Load existing config
		p := config.GetProvider(configName)
		if p != nil {
			m.fields[fieldName].SetValue(configName)
			m.fields[fieldBaseURL].SetValue(p.BaseURL)
			m.fields[fieldAuthToken].SetValue(p.AuthToken)
			m.fields[fieldModel].SetValue(p.Model)
			m.fields[fieldReasoningModel].SetValue(p.ReasoningModel)
			m.fields[fieldHaikuModel].SetValue(p.HaikuModel)
			m.fields[fieldOpusModel].SetValue(p.OpusModel)
			m.fields[fieldSonnetModel].SetValue(p.SonnetModel)
			// Load provider type
			if p.GetType() == config.ProviderTypeOpenAI {
				m.providerType = 1
			}
			// Load env vars for each CLI
			if p.ClaudeEnvVars != nil {
				for k, v := range p.ClaudeEnvVars {
					m.claudeEnvVars[k] = v
				}
			} else if p.EnvVars != nil {
				// Fallback to legacy EnvVars for Claude
				for k, v := range p.EnvVars {
					m.claudeEnvVars[k] = v
				}
			}
			if p.CodexEnvVars != nil {
				for k, v := range p.CodexEnvVars {
					m.codexEnvVars[k] = v
				}
			}
			if p.OpenCodeEnvVars != nil {
				for k, v := range p.OpenCodeEnvVars {
					m.opencodeEnvVars[k] = v
				}
			}
		}
		// Disable name field when editing
		m.focus = fieldType
	} else if presetName != "" {
		// New provider with pre-filled name — skip to next field
		m.fields[fieldName].SetValue(presetName)
		m.focus = fieldType
	} else {
		m.focus = fieldName
	}

	if m.focus != fieldType && m.focus < fieldEnvVars {
		m.fields[m.focus].Focus()
	}
	return m
}

func (m editorModel) init() tea.Cmd {
	return textinput.Blink
}

func (m editorModel) update(msg tea.Msg) (editorModel, tea.Cmd) {
	// After save, ignore everything except saveExitMsg
	if m.saved {
		if _, ok := msg.(saveExitMsg); ok {
			return m, func() tea.Msg { return switchToListMsg{} }
		}
		return m, nil
	}

	// Handle env vars editor mode
	if m.envVarsEdit {
		switch msg := msg.(type) {
		case envVarsExitMsg:
			m.envVarsEdit = false
			// Save to the current CLI's env vars
			switch m.currentEnvCLI {
			case 0:
				m.claudeEnvVars = msg.envVars
			case 1:
				m.codexEnvVars = msg.envVars
			case 2:
				m.opencodeEnvVars = msg.envVars
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.envVarsModel, cmd = m.envVarsModel.update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchToListMsg{} }
		case "tab", "down":
			m.blurCurrentField()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldType
			}
			m.focusCurrentField()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.blurCurrentField()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldEnvVars
			}
			m.focusCurrentField()
			return m, textinput.Blink
		case "ctrl+s", "cmd+s":
			return m.save()
		case "enter":
			if m.focus == fieldType {
				// Toggle type on enter
				m.providerType = (m.providerType + 1) % 2
				return m, nil
			}
			if m.focus == fieldEnvVars {
				// Open env vars editor for current CLI
				var envVars map[string]string
				switch m.currentEnvCLI {
				case 0:
					envVars = m.claudeEnvVars
				case 1:
					envVars = m.codexEnvVars
				case 2:
					envVars = m.opencodeEnvVars
				}
				m.envVarsEdit = true
				m.envVarsModel = newEnvVarsEditorModel(envVars)
				return m, nil
			}
			// Enter on last text field = save
			if m.focus == fieldSonnetModel {
				return m.save()
			}
			// Enter on non-last field = move to next
			m.blurCurrentField()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldType
			}
			m.focusCurrentField()
			return m, textinput.Blink
		case "left", "right":
			// Toggle type with left/right when focused on type field
			if m.focus == fieldType {
				m.providerType = (m.providerType + 1) % 2
				return m, nil
			}
			// Switch CLI with left/right when focused on env vars field
			if m.focus == fieldEnvVars {
				if msg.String() == "left" {
					m.currentEnvCLI = (m.currentEnvCLI - 1 + 3) % 3
				} else {
					m.currentEnvCLI = (m.currentEnvCLI + 1) % 3
				}
				return m, nil
			}
		}
	}

	// Update focused field (only if it's a textinput field)
	if m.focus != fieldType && m.focus < fieldEnvVars {
		var cmd tea.Cmd
		m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *editorModel) blurCurrentField() {
	if m.focus != fieldType && m.focus < fieldEnvVars {
		m.fields[m.focus].Blur()
	}
}

func (m *editorModel) focusCurrentField() {
	if m.focus != fieldType && m.focus < fieldEnvVars {
		m.fields[m.focus].Focus()
	}
}

func (m editorModel) save() (editorModel, tea.Cmd) {
	name := strings.TrimSpace(m.fields[fieldName].Value())
	baseURL := strings.TrimSpace(m.fields[fieldBaseURL].Value())
	token := strings.TrimSpace(m.fields[fieldAuthToken].Value())

	if name == "" {
		m.err = "name is required"
		return m, nil
	}
	if m.editing == "" && config.GetProvider(name) != nil {
		m.err = fmt.Sprintf("provider %q already exists", name)
		return m, nil
	}
	if baseURL == "" {
		m.err = "base URL is required"
		return m, nil
	}
	if token == "" {
		m.err = "auth token is required"
		return m, nil
	}

	// Build ProviderConfig with defaults
	modelDefaults := []struct {
		field editorField
		def   string
	}{
		{fieldModel, "claude-sonnet-4-5"},
		{fieldReasoningModel, "claude-sonnet-4-5"},
		{fieldHaikuModel, "claude-haiku-4-5"},
		{fieldOpusModel, "claude-opus-4-5"},
		{fieldSonnetModel, "claude-sonnet-4-5"},
	}

	modelValues := make([]string, len(modelDefaults))
	for i, md := range modelDefaults {
		val := strings.TrimSpace(m.fields[md.field].Value())
		if val == "" {
			val = md.def
		}
		modelValues[i] = val
	}

	// Determine provider type
	providerType := config.ProviderTypeAnthropic
	if m.providerType == 1 {
		providerType = config.ProviderTypeOpenAI
	}

	p := &config.ProviderConfig{
		Type:           providerType,
		BaseURL:        baseURL,
		AuthToken:      token,
		Model:          modelValues[0],
		ReasoningModel: modelValues[1],
		HaikuModel:     modelValues[2],
		OpusModel:      modelValues[3],
		SonnetModel:    modelValues[4],
	}

	// Add env vars for each CLI if any
	if len(m.claudeEnvVars) > 0 {
		p.ClaudeEnvVars = make(map[string]string)
		for k, v := range m.claudeEnvVars {
			p.ClaudeEnvVars[k] = v
		}
	}
	if len(m.codexEnvVars) > 0 {
		p.CodexEnvVars = make(map[string]string)
		for k, v := range m.codexEnvVars {
			p.CodexEnvVars[k] = v
		}
	}
	if len(m.opencodeEnvVars) > 0 {
		p.OpenCodeEnvVars = make(map[string]string)
		for k, v := range m.opencodeEnvVars {
			p.OpenCodeEnvVars[k] = v
		}
	}

	if err := config.SetProvider(name, p); err != nil {
		m.err = err.Error()
		return m, nil
	}

	if m.editing == "" && m.initMode {
		// First provider — auto-add to default profile
		fbOrder, _ := config.ReadFallbackOrder()
		fbOrder = append(fbOrder, name)
		config.WriteFallbackOrder(fbOrder)
	}

	m.saved = true
	m.status = "Saved"
	m.err = ""
	m.createdName = name
	return m, saveExitTick()
}

func (m editorModel) view(width, height int) string {
	// If editing env vars, show that view
	if m.envVarsEdit {
		return m.envVarsModel.view(width, height)
	}

	// Use global layout dimensions
	contentWidth, _, _, _ := LayoutDimensions(width, height)
	sidePadding := 2

	var b strings.Builder

	// Header
	title := "Add Provider"
	icon := "➕"
	if m.editing != "" {
		title = fmt.Sprintf("Edit Provider: %s", m.editing)
		icon = "✏️"
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render(icon + " " + title)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Form content
	var content strings.Builder
	content.WriteString(sectionTitleStyle.Render(" Provider Settings"))
	content.WriteString("\n\n")

	for i := range m.fields {
		if editorField(i) == fieldType {
			// Special handling for type field
			cursor := "  "
			style := dimStyle
			if m.focus == fieldType {
				cursor = "▸ "
				style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
			}
			typeLabel := "Anthropic Messages API"
			if m.providerType == 1 {
				typeLabel = "OpenAI Chat Completions API"
			}
			content.WriteString(style.Render(fmt.Sprintf("%sAPI Type:         [%s] (←/→ to change)", cursor, typeLabel)))
			content.WriteString("\n")
			continue
		}
		if editorField(i) == fieldEnvVars {
			// Special handling for env vars field with CLI tabs
			cursor := "  "
			style := dimStyle
			if m.focus == fieldEnvVars {
				cursor = "▸ "
				style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
			}

			// Get current CLI name and env var count
			cliNames := []string{"Claude", "Codex", "OpenCode"}
			cliName := cliNames[m.currentEnvCLI]
			var envCount int
			switch m.currentEnvCLI {
			case 0:
				envCount = len(m.claudeEnvVars)
			case 1:
				envCount = len(m.codexEnvVars)
			case 2:
				envCount = len(m.opencodeEnvVars)
			}

			envLabel := fmt.Sprintf("%d configured", envCount)
			if envCount == 0 {
				envLabel = "none"
			}

			// Show CLI selector and env var count
			content.WriteString(style.Render(fmt.Sprintf("%sEnv Vars (%s): [%s] (←/→ CLI, enter edit)", cursor, cliName, envLabel)))
			content.WriteString("\n")
			continue
		}
		if m.editing != "" && editorField(i) == fieldName {
			content.WriteString(dimStyle.Render(fmt.Sprintf("  Name:             %s", m.editing)))
			content.WriteString("\n")
			continue
		}
		content.WriteString(m.fields[i].View())
		content.WriteString("\n")
	}

	// Form box with proper width
	formWidth := contentWidth * 75 / 100
	if formWidth < 60 {
		formWidth = 60
	}
	if formWidth > 100 {
		formWidth = 100
	}

	formBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Width(formWidth).
		Padding(0, 1).
		Render(content.String())
	b.WriteString(formBox)

	b.WriteString("\n\n")
	if m.saved {
		b.WriteString(successStyle.Render("  ✓ " + m.status))
	} else if m.err != "" {
		errBox := lipgloss.NewStyle().
			Foreground(errorColor).
			Render("✗ " + m.err)
		b.WriteString(errBox)
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
	helpBar := RenderHelpBar("Tab next • "+saveKeyHint()+" save • Esc cancel", width)
	view.WriteString(helpBar)

	return view.String()
}

// standaloneEditorModel wraps editorModel for standalone use.
type standaloneEditorModel struct {
	editor      editorModel
	cancelled   bool
	createdName string
}

func newStandaloneEditorModel(configName string) standaloneEditorModel {
	return standaloneEditorModel{
		editor: newEditorModel(configName),
	}
}

func (m standaloneEditorModel) Init() tea.Cmd {
	return m.editor.init()
}

func (m standaloneEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
		if msg.String() == "esc" && !m.editor.saved {
			m.cancelled = true
			return m, tea.Quit
		}
	case switchToListMsg:
		// Editor finished saving — extract the name and quit
		m.createdName = strings.TrimSpace(m.editor.fields[fieldName].Value())
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.update(msg)
	return m, cmd
}

func (m standaloneEditorModel) View() string {
	return m.editor.view(0, 0)
}

// RunAddProvider runs a standalone provider editor TUI for creating a new provider.
// If presetName is non-empty, it pre-fills the name field. If that name already exists,
// an error is returned immediately.
// After saving, if profiles exist, it runs a profile multi-select so the user can
// choose which profiles to add the new provider to.
func RunAddProvider(presetName string) (string, error) {
	if presetName != "" && config.GetProvider(presetName) != nil {
		return "", fmt.Errorf("provider %q already exists", presetName)
	}
	m := standaloneEditorModel{
		editor: newEditorModelWithPreset("", presetName),
	}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	sm := result.(standaloneEditorModel)
	if sm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	providerName := sm.createdName

	// If profiles exist, let the user pick which ones to add this provider to
	profiles := config.ListProfiles()
	if len(profiles) > 0 {
		selected, err := RunProfileMultiSelect()
		if err != nil {
			// cancelled or error — provider is already saved, just skip profile assignment
			return providerName, nil
		}
		for _, profile := range selected {
			order, _ := config.ReadProfileOrder(profile)
			order = append(order, providerName)
			config.WriteProfileOrder(profile, order)
		}
	}

	return providerName, nil
}

// RunEditProvider runs a standalone provider editor TUI for editing an existing provider.
func RunEditProvider(name string) error {
	m := newStandaloneEditorModel(name)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	sm := result.(standaloneEditorModel)
	if sm.cancelled {
		return fmt.Errorf("cancelled")
	}
	return nil
}

// envVarsExitMsg signals that the env vars editor is done.
type envVarsExitMsg struct {
	envVars map[string]string
}

// envVarsEditorModel is a sub-editor for managing key-value environment variables.
type envVarsEditorModel struct {
	entries     []envVarEntry
	cursor      int
	phase       int // 0=list, 1=edit key, 2=edit value
	keyInput    string
	valueInput  string
	editingIdx  int // index being edited, -1 for new
}

type envVarEntry struct {
	key   string
	value string
}

func newEnvVarsEditorModel(envVars map[string]string) envVarsEditorModel {
	var entries []envVarEntry
	for k, v := range envVars {
		entries = append(entries, envVarEntry{key: k, value: v})
	}
	// Sort entries by key for consistent display
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].key > entries[j].key {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return envVarsEditorModel{
		entries:    entries,
		editingIdx: -1,
	}
}

func (m envVarsEditorModel) update(msg tea.Msg) (envVarsEditorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.phase == 1 {
			// Editing key
			return m.updateKeyEdit(msg)
		}
		if m.phase == 2 {
			// Editing value
			return m.updateValueEdit(msg)
		}
		// Phase 0: list view
		return m.updateList(msg)
	}
	return m, nil
}

func (m envVarsEditorModel) updateList(msg tea.KeyMsg) (envVarsEditorModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		// Return to provider editor with current env vars
		envVars := make(map[string]string)
		for _, e := range m.entries {
			if e.key != "" && e.value != "" {
				envVars[e.key] = e.value
			}
		}
		return m, func() tea.Msg { return envVarsExitMsg{envVars: envVars} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries) {
			m.cursor++
		}
	case "enter":
		if m.cursor == len(m.entries) {
			// Add new entry
			m.phase = 1
			m.editingIdx = -1
			m.keyInput = ""
			m.valueInput = ""
		} else {
			// Edit existing entry
			m.phase = 1
			m.editingIdx = m.cursor
			m.keyInput = m.entries[m.cursor].key
			m.valueInput = m.entries[m.cursor].value
		}
	case "d", "x":
		// Delete entry
		if m.cursor < len(m.entries) {
			m.entries = append(m.entries[:m.cursor], m.entries[m.cursor+1:]...)
			if m.cursor >= len(m.entries) && m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

func (m envVarsEditorModel) updateKeyEdit(msg tea.KeyMsg) (envVarsEditorModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.phase = 0
		m.keyInput = ""
		m.valueInput = ""
	case "enter", "tab":
		if m.keyInput != "" {
			m.phase = 2 // Move to value editing
		}
	case "backspace":
		if len(m.keyInput) > 0 {
			m.keyInput = m.keyInput[:len(m.keyInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.keyInput += msg.String()
		}
	}
	return m, nil
}

func (m envVarsEditorModel) updateValueEdit(msg tea.KeyMsg) (envVarsEditorModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.phase = 0
		m.keyInput = ""
		m.valueInput = ""
	case "enter":
		// Save the entry
		if m.keyInput != "" {
			entry := envVarEntry{key: m.keyInput, value: m.valueInput}
			if m.editingIdx >= 0 {
				m.entries[m.editingIdx] = entry
			} else {
				m.entries = append(m.entries, entry)
				m.cursor = len(m.entries) - 1
			}
		}
		m.phase = 0
		m.keyInput = ""
		m.valueInput = ""
	case "backspace":
		if len(m.valueInput) > 0 {
			m.valueInput = m.valueInput[:len(m.valueInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.valueInput += msg.String()
		}
	}
	return m, nil
}

func (m envVarsEditorModel) view(width, height int) string {
	// Use global layout dimensions
	contentWidth, _, _, _ := LayoutDimensions(width, height)
	sidePadding := 2

	var b strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(headerBgColor).
		Padding(0, 2).
		Render("🔧 Environment Variables")
	b.WriteString(header)
	b.WriteString("\n\n")

	var content strings.Builder

	if m.phase == 1 {
		// Key editing
		content.WriteString(sectionTitleStyle.Render(" Enter Variable Name"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" e.g. CLAUDE_CODE_MAX_OUTPUT_TOKENS"))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("  " + m.keyInput + "█"))
	} else if m.phase == 2 {
		// Value editing
		content.WriteString(sectionTitleStyle.Render(fmt.Sprintf(" Enter Value for %s", m.keyInput)))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(accentColor).Render("  " + m.valueInput + "█"))
	} else {
		// List view
		content.WriteString(sectionTitleStyle.Render(" Custom Environment Variables"))
		content.WriteString("\n")
		content.WriteString(dimStyle.Render(" These are passed as x-env-* headers to the proxy"))
		content.WriteString("\n\n")

		for i, e := range m.entries {
			cursor := "  "
			style := tableRowStyle
			if i == m.cursor {
				cursor = "▸ "
				style = tableSelectedRowStyle
			}
			line := fmt.Sprintf("%s%s = %s", cursor, e.key, e.value)
			content.WriteString(style.Render(line))
			content.WriteString("\n")
		}

		// Add new entry option
		cursor := "  "
		style := dimStyle
		if m.cursor == len(m.entries) {
			cursor = "▸ "
			style = tableSelectedRowStyle
		}
		content.WriteString(style.Render(cursor + "[+ Add new variable]"))
	}

	// Content box with proper width
	boxWidth := contentWidth * 60 / 100
	if boxWidth < 60 {
		boxWidth = 60
	}
	if boxWidth > 80 {
		boxWidth = 80
	}

	contentBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(boxWidth).
		Render(content.String())
	b.WriteString(contentBox)

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
	var helpText string
	if m.phase == 1 {
		helpText = "Enter/Tab next • Esc cancel"
	} else if m.phase == 2 {
		helpText = "Enter save • Esc cancel"
	} else {
		helpText = "↑↓ move • Enter edit/add • d delete • Esc done"
	}
	helpBar := RenderHelpBar(helpText, width)
	view.WriteString(helpBar)

	return view.String()
}
