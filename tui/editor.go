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
	fieldBaseURL
	fieldAuthToken
	fieldModel
	fieldReasoningModel
	fieldHaikuModel
	fieldOpusModel
	fieldSonnetModel
	fieldCount
)

type editorModel struct {
	fields      [fieldCount]textinput.Model
	focus       editorField
	editing     string // config name being edited, empty = new
	initMode    bool   // true = auto-add to default profile on save (first provider)
	err         string
	saved       bool   // true = save succeeded, waiting to exit
	status      string // "Saved" success message
	createdName string // name of provider after save (for callers)
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

	m := editorModel{
		fields:  fields,
		editing: configName,
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
		}
		// Disable name field when editing
		m.focus = fieldBaseURL
	} else if presetName != "" {
		// New provider with pre-filled name — skip to next field
		m.fields[fieldName].SetValue(presetName)
		m.focus = fieldBaseURL
	} else {
		m.focus = fieldName
	}

	m.fields[m.focus].Focus()
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchToListMsg{} }
		case "tab", "down":
			m.fields[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldBaseURL
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.fields[m.focus].Blur()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldSonnetModel
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		case "ctrl+s", "cmd+s", "enter":
			isSaveKey := (isMac && msg.String() == "cmd+s") || (!isMac && msg.String() == "ctrl+s")
			if m.focus == fieldCount-1 || isSaveKey {
				return m.save()
			}
			// Enter on non-last field = move to next
			m.fields[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			if m.editing != "" && m.focus == fieldName {
				m.focus = fieldBaseURL
			}
			m.fields[m.focus].Focus()
			return m, textinput.Blink
		}
	}

	// Update focused field
	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
	return m, cmd
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

	p := &config.ProviderConfig{
		BaseURL:        baseURL,
		AuthToken:      token,
		Model:          modelValues[0],
		ReasoningModel: modelValues[1],
		HaikuModel:     modelValues[2],
		OpusModel:      modelValues[3],
		SonnetModel:    modelValues[4],
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
		if m.editing != "" && editorField(i) == fieldName {
			content.WriteString(dimStyle.Render(fmt.Sprintf("  Name:             %s", m.editing)))
			content.WriteString("\n")
			continue
		}
		content.WriteString(m.fields[i].View())
		content.WriteString("\n")
	}

	formBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
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
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  tab next • " + saveKeyHint() + " save • esc cancel"))
	} else {
		b.WriteString(helpStyle.Render("  tab next • " + saveKeyHint() + " save • esc cancel"))
	}

	return b.String()
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
