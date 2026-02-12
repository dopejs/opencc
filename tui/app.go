package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dopejs/opencc/internal/config"
)

type view int

const (
	viewList view = iota
	viewEditor
	viewFallback
	viewProfileList
	viewRouting
)

type model struct {
	currentView view
	list        listModel
	editor      editorModel
	fallback    fallbackModel
	profileList profileListModel
	routing     routingModel
	width       int
	height      int
	err         error
}

func initialModel() model {
	return model{
		currentView: viewList,
		list:        newListModel(),
		fallback:    newFallbackModel("default"),
	}
}

func (m model) Init() tea.Cmd {
	return m.list.init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.currentView {
	case viewList:
		return m.updateList(msg)
	case viewEditor:
		return m.updateEditor(msg)
	case viewFallback:
		return m.updateFallback(msg)
	case viewProfileList:
		return m.updateProfileList(msg)
	case viewRouting:
		return m.updateRouting(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case viewList:
		return m.list.view(m.width, m.height)
	case viewEditor:
		return m.editor.view(m.width, m.height)
	case viewFallback:
		return m.fallback.view(m.width, m.height)
	case viewProfileList:
		return m.profileList.view(m.width, m.height)
	case viewRouting:
		return m.routing.view(m.width, m.height)
	}
	return ""
}

// Run starts the TUI application.
func Run() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// RunPick starts a standalone checkbox picker TUI and returns selected provider names.
func RunPick() ([]string, error) {
	m := newPickModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	pm := result.(pickModel)
	if pm.cancelled {
		return nil, fmt.Errorf("cancelled")
	}
	return pm.Result(), nil
}

// RunCreateFirst starts a standalone editor TUI for creating the first provider.
// Returns the created provider name.
func RunCreateFirst() (string, error) {
	m := newCreateFirstModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	cm := result.(createFirstModel)
	if cm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return cm.createdName, nil
}

// createFirstModel wraps the editor for standalone first-provider creation.
type createFirstModel struct {
	editor    editorModel
	cancelled bool
	createdName string
	width     int
	height    int
}

func newCreateFirstModel() createFirstModel {
	editor := newEditorModel("")
	editor.initMode = true
	return createFirstModel{
		editor: editor,
	}
}

func (m createFirstModel) Init() tea.Cmd {
	return m.editor.init()
}

func (m createFirstModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		m.createdName = m.editor.fields[fieldName].Value()
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.update(msg)
	return m, cmd
}

func (m createFirstModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("  No providers configured. Create one to get started:"))
	b.WriteString("\n\n")
	b.WriteString(m.editor.view(m.width, m.height))
	return b.String()
}

// Messages
type switchToListMsg struct{}
type switchToEditorMsg struct {
	configName string // empty = new config
}
type switchToFallbackMsg struct {
	profile string
}
type switchToProfileListMsg struct{}
type statusMsg struct {
	text string
}
type saveExitMsg struct{}

// saveExitDelay is the duration to show the "Saved" feedback before exiting.
const saveExitDelay = 500 * time.Millisecond

func saveExitTick() tea.Cmd {
	return tea.Tick(saveExitDelay, func(time.Time) tea.Msg {
		return saveExitMsg{}
	})
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchToEditorMsg:
		m.currentView = viewEditor
		m.editor = newEditorModel(msg.configName)
		return m, m.editor.init()
	case switchToFallbackMsg:
		m.currentView = viewFallback
		m.fallback = newFallbackModel(msg.profile)
		return m, m.fallback.init()
	case switchToProfileListMsg:
		m.currentView = viewProfileList
		m.profileList = newProfileListModel()
		return m, m.profileList.init()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.update(msg)
	return m, cmd
}

func (m model) updateEditor(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case switchToListMsg:
		m.currentView = viewList
		m.list = newListModel()
		return m, m.list.init()
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.update(msg)
	return m, cmd
}

func (m model) updateFallback(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchToListMsg:
		// Return to profile list (fallback is always entered from there)
		m.currentView = viewProfileList
		m.profileList = newProfileListModel()
		return m, m.profileList.init()
	case switchToRoutingMsg:
		m.currentView = viewRouting
		m.routing = newRoutingModel(msg.profile)
		return m, m.routing.init()
	}

	var cmd tea.Cmd
	m.fallback, cmd = m.fallback.update(msg)
	return m, cmd
}

func (m model) updateRouting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case switchToListMsg:
		// Return to profile list
		m.currentView = viewProfileList
		m.profileList = newProfileListModel()
		return m, m.profileList.init()
	}

	var cmd tea.Cmd
	m.routing, cmd = m.routing.update(msg)
	return m, cmd
}

func (m model) updateProfileList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case switchToListMsg:
		m.currentView = viewList
		m.list = newListModel()
		return m, m.list.init()
	case switchToFallbackMsg:
		m.currentView = viewFallback
		m.fallback = newFallbackModel(msg.profile)
		return m, m.fallback.init()
	}

	var cmd tea.Cmd
	m.profileList, cmd = m.profileList.update(msg)
	return m, cmd
}

// RunProfilePicker starts a standalone profile picker TUI and returns the selected profile name.
func RunProfilePicker() (string, error) {
	m := newProfilePickerModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	pm := result.(profilePickerModel)
	if pm.cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return pm.selected, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RunSelectType runs a selector TUI to choose between "provider" and "group".
func RunSelectType() (string, error) {
	items := []selectorItem{
		{name: "provider"},
		{name: "group"},
	}
	return RunSelector("Select type", items)
}

// RunSelectProvider runs a selector TUI to choose a provider.
// minCount specifies the minimum number of providers required (for delete validation).
func RunSelectProvider(minCount int) (string, error) {
	names := config.ProviderNames()
	if len(names) == 0 {
		return "", fmt.Errorf("no providers configured")
	}

	items := make([]selectorItem, len(names))
	for i, name := range names {
		items[i] = selectorItem{name: name}
		if minCount > 0 && len(names) <= minCount {
			items[i].disabled = true
			items[i].reason = "cannot delete last provider"
		}
	}
	return RunSelector("Select provider", items)
}

// RunSelectGroup runs a selector TUI to choose a group.
// excludeDefault prevents selecting the "default" group.
func RunSelectGroup(excludeDefault bool) (string, error) {
	names := config.ListProfiles()
	if len(names) == 0 {
		return "", fmt.Errorf("no groups configured")
	}

	items := make([]selectorItem, len(names))
	for i, name := range names {
		items[i] = selectorItem{name: name}
		if excludeDefault && name == "default" {
			items[i].disabled = true
			items[i].reason = "cannot delete default"
		}
	}
	return RunSelector("Select group", items)
}

// RunAddGroup runs the group creation flow: name input then group editor.
// If presetName is non-empty, it pre-fills the name field.
func RunAddGroup(presetName string) error {
	name, err := RunGroupCreate(presetName)
	if err != nil {
		return err
	}
	return RunEditGroup(name)
}

// RunEditGroup runs a standalone group editor TUI.
func RunEditGroup(name string) error {
	return RunEditProfile(name)
}

