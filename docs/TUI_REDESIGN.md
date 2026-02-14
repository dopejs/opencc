# TUI Redesign Plan

## Current State Analysis

### Data Model Summary

```
OpenCCConfig
├── Providers: map[string]*ProviderConfig
│   ├── BaseURL (required)
│   ├── AuthToken (required)
│   ├── Model
│   ├── ReasoningModel
│   ├── HaikuModel
│   ├── OpusModel
│   ├── SonnetModel
│   └── EnvVars: map[string]string
│
├── Profiles: map[string]*ProfileConfig
│   ├── Providers: []string (ordered list)
│   ├── LongContextThreshold: int
│   └── Routing: map[Scenario]*ScenarioRoute
│       └── ScenarioRoute
│           └── Providers: []*ProviderRoute
│               ├── Name
│               └── Model (override)
│
└── ProjectBindings: map[string]string (path -> profile)
```

### Current TUI Screens

| Screen | Purpose | Fields/Actions |
|--------|---------|----------------|
| List | Provider list | add, edit, delete, go to profiles |
| Editor | Provider form | 8 text fields + env vars sub-editor |
| ProfileList | Profile list | add, edit, delete profiles |
| Fallback | Profile editor | default providers (reorder) + scenario routing |
| Routing | Scenario editor | 5 scenarios, provider select + model override |
| EnvVars | Key-value editor | list, add, edit, delete env vars |

### Current Navigation Flow

```
List ─────────────────────────────────────────────────────────────┐
  │                                                               │
  ├─[a]─> Editor (new) ─[save]─> ProfileMultiSelect ─[save]──────>│
  │                                                               │
  ├─[e]─> Editor (edit) ─[save]─────────────────────────────────>│
  │                                                               │
  └─[f]─> ProfileList ────────────────────────────────────────────┤
            │                                                     │
            ├─[a]─> GroupCreate ─> Fallback ─[save]──────────────>│
            │                                                     │
            └─[enter]─> Fallback ─────────────────────────────────┤
                          │                                       │
                          └─[enter on scenario]─> ScenarioEdit ──>│
```

### Current Pain Points

1. **Deep nesting**: Provider → Profile → Scenario → Model override (4 levels)
2. **Context loss**: Hard to see overall config while editing details
3. **Repetitive navigation**: Must go back and forth to compare settings
4. **Growing complexity**: v1.5.0 will add provider type, CLI selection

---

## v1.5.0 New Requirements

### Provider Type (Request Transform)

New field in ProviderConfig:
```go
type ProviderConfig struct {
    Type string `json:"type,omitempty"` // "anthropic", "openai", "azure", "bedrock"
    // ... existing fields
}
```

### CLI Selection & Default Profile

New top-level config fields:
```go
type OpenCCConfig struct {
    CLI            string `json:"cli,omitempty"`             // "claude", "codex", "opencode"
    DefaultProfile string `json:"default_profile,omitempty"` // default profile name
    WebPort        int    `json:"web_port,omitempty"`        // Web UI port (default 19841)
    // ... existing fields
}
```

Or CLI per-profile:
```go
type ProfileConfig struct {
    CLI string `json:"cli,omitempty"`
    // ... existing fields
}
```

### Default Profile Logic Changes

当前代码硬编码了名为 "default" 的 profile 不能删除。引入 `DefaultProfile` 设置后需要修改：

1. **删除保护**: 不能删除的是 `config.DefaultProfile` 指向的 profile，而不是名为 "default" 的
2. **启动时读取**: `cmd/root.go` 中读取 `config.DefaultProfile`，如果未设置则回退到 "default"
3. **向后兼容**: 如果 `DefaultProfile` 为空，行为与现在一致（使用名为 "default" 的 profile）

需要修改的文件：
- `internal/config/config.go` - 添加 `DefaultProfile` 和 `WebPort` 字段
- `internal/config/store.go:244` - 删除逻辑改为检查 `DefaultProfile`
- `cmd/root.go:397-399` - 读取 `DefaultProfile` 而不是硬编码 "default"
- `cmd/root.go:437` - 写入时使用 `DefaultProfile`
- `cmd/bind.go:107` - 提示信息使用实际的 default profile 名
- `tui/profile_list.go:96-97` - 检查 `DefaultProfile`
- `tui/config_main.go:111-112` - 同上
- `internal/web/api_profiles.go:214-215` - Web API 同上

---

## Redesign Options

### Option A: Hierarchical Tree Navigation

```
┌─ OpenCC Config ─────────────────────────────────────────────────┐
│                                                                 │
│  ▼ Providers                                                    │
│    ├─ anthropic-direct                                          │
│    ├─ openrouter                                                │
│    └─ [+ Add Provider]                                          │
│                                                                 │
│  ▼ Profiles                                                     │
│    ├─ default                                                   │
│    │   ├─ Providers: anthropic-direct, openrouter               │
│    │   └─ Routing: 2 scenarios configured                       │
│    ├─ work                                                      │
│    └─ [+ Add Profile]                                           │
│                                                                 │
│  ▶ Project Bindings (3)                                         │
│                                                                 │
│  ▶ Settings                                                     │
│     └─ CLI: claude                                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

Pros:
- See everything at a glance
- Expand/collapse sections
- Natural hierarchy

Cons:
- Complex tree state management
- May feel cramped on small terminals

### Option B: Tab-Based Sections

```
┌─ OpenCC ────────────────────────────────────────────────────────┐
│  [Providers]  [Profiles]  [Bindings]  [Settings]                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Providers                                                      │
│  ──────────                                                     │
│  > anthropic-direct    api.anthropic.com     anthropic          │
│    openrouter          openrouter.ai/api     openai             │
│    azure-prod          azure.openai.com      azure              │
│                                                                 │
│  [a] Add  [e] Edit  [d] Delete                                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

Pros:
- Clear separation of concerns
- Familiar UI pattern
- Easy to add new tabs

Cons:
- Can't see cross-section relationships
- Tab switching overhead

### Option C: Dashboard + Detail Panes

```
┌─ Dashboard ─────────────────────┬─ Detail ──────────────────────┐
│                                 │                               │
│  Providers (3)                  │  Provider: anthropic-direct   │
│  > anthropic-direct ────────────│  ─────────────────────────    │
│    openrouter                   │  Type:     anthropic          │
│    azure-prod                   │  Base URL: api.anthropic.com  │
│                                 │  Model:    claude-sonnet-4-5  │
│  Profiles (2)                   │                               │
│    default                      │  Used in profiles:            │
│    work                         │  - default (primary)          │
│                                 │  - work (fallback #2)         │
│  Bindings (3)                   │                               │
│    ~/code/proj1 → default       │  [e] Edit  [d] Delete         │
│                                 │                               │
└─────────────────────────────────┴───────────────────────────────┘
```

Pros:
- See relationships (which profiles use this provider)
- Quick preview without entering edit mode
- Good information density

Cons:
- Requires wider terminal
- Split-pane complexity

### Option D: Simplified TUI + Web UI for Complex Config

Keep TUI simple:
- Quick provider/profile selection
- Basic add/edit/delete
- Launch web UI for complex config

```
┌─ OpenCC ────────────────────────────────────────────────────────┐
│                                                                 │
│  Quick Actions                                                  │
│  > Switch Profile                                               │
│    Add Provider                                                 │
│    Edit Provider                                                │
│    Open Web UI (for advanced config)                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

Pros:
- TUI stays simple and fast
- Complex config in proper UI
- Less code to maintain

Cons:
- Requires browser
- Two UIs to maintain

---

## Recommended: Main Menu + Dashboard + Detail

结合主菜单和分栏布局，支持系统级设置和多 CLI。

### Overall Navigation Structure

```
┌─────────────────────────────────────────────────────────────────┐
│                         Main Menu                               │
│                                                                 │
│                    ┌─────────────────┐                          │
│                    │    OpenCC       │                          │
│                    ├─────────────────┤                          │
│                    │ > Launch        │ ─── Quick start CLI      │
│                    │   Configure     │ ─── Dashboard view       │
│                    │   Settings      │ ─── Global settings      │
│                    │   Web UI        │ ─── Open browser         │
│                    │   Quit          │                          │
│                    └─────────────────┘                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    ┌──────────┐      ┌─────────────┐      ┌──────────┐
    │  Launch  │      │  Configure  │      │ Settings │
    │  Wizard  │      │  Dashboard  │      │   Form   │
    └──────────┘      └─────────────┘      └──────────┘
```

### Main Menu Screen

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│                           OpenCC                                │
│                    Environment Switcher                         │
│                                                                 │
│                    ┌─────────────────────┐                      │
│                    │                     │                      │
│                    │  >  Launch          │                      │
│                    │     Configure       │                      │
│                    │     Settings        │                      │
│                    │     Web UI          │                      │
│                    │     Quit            │                      │
│                    │                     │                      │
│                    └─────────────────────┘                      │
│                                                                 │
│              Current: default profile, claude CLI               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Launch Wizard (Quick Start)

快速选择 Profile 和 CLI 启动：

```
┌─────────────────────────────────────────────────────────────────┐
│ Launch                                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Select Profile:                                                │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ > default      anthropic-direct, openrouter             │    │
│  │   work         azure-prod                               │    │
│  │   personal     anthropic-direct                         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  Select CLI:                                                    │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ > claude       Claude Code (Anthropic)                  │    │
│  │   codex        Codex CLI (OpenAI)                       │    │
│  │   opencode     OpenCode                                 │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│                                        [Enter] Launch  [Esc] Back│
└─────────────────────────────────────────────────────────────────┘
```

### Settings Screen (Global)

系统级设置：

```
┌─────────────────────────────────────────────────────────────────┐
│ Settings                                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ─── Defaults ───                                               │
│  Default CLI:          [claude       ▼]                         │
│  Default Profile:      [default      ▼]                         │
│                                                                 │
│  ─── Web UI ───                                                 │
│  Port:                 [19841________]                          │
│                                                                 │
│  ─── Advanced ───                                               │
│  Config path:          ~/.opencc/opencc.json                    │
│  [Reset to defaults]                                            │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ [Tab] next  [Shift+Tab] prev  [Ctrl+S] save  [Esc] back         │
└─────────────────────────────────────────────────────────────────┘
```

注：Proxy 使用随机端口，随 CLI 进程自动启动，无需用户配置。

### Configure Dashboard (Split View)

配置管理的分栏视图：

```
┌─ Configure ─────────────────────┬───────────────────────────────┐
│                                 │                               │
│ ▼ Providers (3)                 │ Provider: anthropic-direct    │
│   > anthropic-direct            │ ─────────────────────────     │
│     openrouter                  │ Type:     anthropic           │
│     azure-prod                  │ Base URL: api.anthropic.com   │
│                                 │ Auth:     sk-ant-***...xyz    │
│ ▼ Profiles (2)                  │                               │
│     default                     │ Models:                       │
│     work                        │   Default:   claude-sonnet-4-5│
│                                 │   Reasoning: claude-sonnet-4-5│
│ ▶ Project Bindings (3)          │   Haiku:     claude-haiku-4-5 │
│                                 │   Opus:      claude-opus-4-5  │
│                                 │                               │
│                                 │ Env Vars: 2 configured        │
│                                 │                               │
│                                 │ Used in profiles:             │
│                                 │   default (primary)           │
│                                 │   work (fallback #2)          │
│                                 │                               │
├─────────────────────────────────┴───────────────────────────────┤
│ [a]dd [e]dit [d]elete [/] search [Esc] menu [q]uit              │
└─────────────────────────────────────────────────────────────────┘
```

### Proposed Component Architecture

```
tui/
├── app.go                # Main app, routes between screens
├── menu.go               # Main menu model
├── launch.go             # Launch wizard model
├── settings.go           # Global settings form
├── configure/
│   ├── dashboard.go      # Split-view container
│   ├── sidebar.go        # Left pane (collapsible sections)
│   └── detail.go         # Right pane (item details)
├── components/
│   ├── list.go           # Generic scrollable list
│   ├── form.go           # Generic form with fields
│   ├── select.go         # Dropdown/single select
│   ├── multiselect.go    # Checkbox multi-select with reorder
│   ├── keyvalue.go       # Key-value editor
│   ├── modal.go          # Modal/overlay container
│   └── help.go           # Help bar component
├── views/
│   ├── provider.go       # Provider detail + edit form
│   ├── profile.go        # Profile detail + edit form
│   ├── binding.go        # Binding detail + edit
│   └── scenario.go       # Scenario routing editor
└── styles.go             # Shared styles
```

### Component Specifications

#### 1. List Component (`components/list.go`)

Generic scrollable list with:
- Cursor navigation (j/k, up/down)
- Selection callback
- Optional: inline actions (d for delete)
- Optional: section headers

```go
type ListItem struct {
    ID       string
    Label    string
    Sublabel string
    Icon     string // optional prefix
}

type ListModel struct {
    items    []ListItem
    cursor   int
    selected string
    onSelect func(id string) tea.Cmd
}
```

#### 2. Form Component (`components/form.go`)

Generic form with:
- Multiple field types (text, select, toggle)
- Tab navigation between fields
- Validation
- Save/cancel actions

```go
type FieldType int
const (
    FieldText FieldType = iota
    FieldSelect
    FieldToggle
    FieldKeyValue // opens sub-editor
)

type Field struct {
    Key         string
    Label       string
    Type        FieldType
    Value       string
    Options     []string  // for FieldSelect
    Required    bool
    Placeholder string
}

type FormModel struct {
    fields  []Field
    focused int
    onSave  func(values map[string]string) tea.Cmd
}
```

#### 3. Dashboard Model (`dashboard.go`)

Left pane with collapsible sections:

```go
type Section struct {
    Name      string
    Collapsed bool
    Items     []ListItem
}

type DashboardModel struct {
    sections []Section
    cursor   struct {
        section int
        item    int
    }
}
```

#### 4. Detail Model (`detail.go`)

Right pane showing selected item details:

```go
type DetailModel struct {
    itemType string // "provider", "profile", "binding"
    itemID   string
    content  string // rendered detail view
    actions  []Action
}

type Action struct {
    Key   string
    Label string
    Cmd   func() tea.Cmd
}
```

### Screen Layouts

#### Main Dashboard

```
┌─────────────────────────────────┬───────────────────────────────┐
│ ▼ Providers (3)                 │ anthropic-direct              │
│   > anthropic-direct            │ ───────────────────────────── │
│     openrouter                  │ Type:     anthropic           │
│     azure-prod                  │ Base URL: api.anthropic.com   │
│                                 │ Auth:     ****...abc          │
│ ▼ Profiles (2)                  │ Model:    claude-sonnet-4-5   │
│     default                     │                               │
│     work                        │ Models:                       │
│                                 │   Reasoning: claude-sonnet-4-5│
│ ▶ Bindings (3)                  │   Haiku: claude-haiku-4-5     │
│                                 │   Opus: claude-opus-4-5       │
│ ▶ Settings                      │   Sonnet: claude-sonnet-4-5   │
│                                 │                               │
│                                 │ Env Vars: 2 configured        │
│                                 │                               │
│                                 │ Used in:                      │
│                                 │   default (primary)           │
│                                 │   work (fallback #2)          │
├─────────────────────────────────┴───────────────────────────────┤
│ [a]dd [e]dit [d]elete [enter] expand  [q]uit                    │
└─────────────────────────────────────────────────────────────────┘
```

#### Provider Edit (Modal/Overlay)

```
┌─────────────────────────────────────────────────────────────────┐
│ Edit Provider: anthropic-direct                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Type:            [anthropic    ▼]                              │
│  Base URL:        [https://api.anthropic.com________________]   │
│  Auth Token:      [sk-ant-************************************] │
│                                                                 │
│  ─── Models ───                                                 │
│  Default:         [claude-sonnet-4-5________________________]   │
│  Reasoning:       [claude-sonnet-4-5________________________]   │
│  Haiku:           [claude-haiku-4-5_________________________]   │
│  Opus:            [claude-opus-4-5__________________________]   │
│  Sonnet:          [claude-sonnet-4-5________________________]   │
│                                                                 │
│  ─── Environment Variables ───                                  │
│  [Edit Env Vars...]                                             │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ [Tab] next field  [Shift+Tab] prev  [Ctrl+S] save  [Esc] cancel │
└─────────────────────────────────────────────────────────────────┘
```

#### Profile Edit (Modal/Overlay)

```
┌─────────────────────────────────────────────────────────────────┐
│ Edit Profile: default                                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  CLI:                    [claude       ▼]                       │
│  Long Context Threshold: [32000_________] tokens                │
│                                                                 │
│  ─── Default Providers (drag to reorder) ───                    │
│  1. [x] anthropic-direct                                        │
│  2. [x] openrouter                                              │
│  3. [ ] azure-prod                                              │
│                                                                 │
│  ─── Scenario Routing ───                                       │
│  > think        anthropic-direct → claude-sonnet-4-5            │
│    image        (use default)                                   │
│    longContext  openrouter → claude-sonnet-4-5                  │
│    webSearch    (use default)                                   │
│    background   (use default)                                   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ [Space] toggle  [Enter] edit scenario  [Ctrl+S] save  [Esc] back│
└─────────────────────────────────────────────────────────────────┘
```

---

## Implementation Plan

### Phase 1: Core Components

1. `components/list.go` - Generic list with cursor
2. `components/form.go` - Generic form with fields
3. `components/keyvalue.go` - Key-value editor

### Phase 2: Layout

1. `dashboard.go` - Left pane with sections
2. `detail.go` - Right pane with item details
3. `app.go` - Split layout management

### Phase 3: Views

1. `views/providers.go` - Provider list + edit form
2. `views/profiles.go` - Profile list + edit form
3. `views/bindings.go` - Bindings list
4. `views/settings.go` - Global settings

### Phase 4: Migration

1. Keep old TUI working during development
2. Add feature flag to switch between old/new
3. Remove old TUI after testing

---

## Exit Code Convention

程序应尽量避免 exit code > 0，除非是真正的程序错误。以下情况应该 exit 0 并给出友好提示：

| 场景 | Exit Code | 用户提示 |
|------|-----------|----------|
| 用户取消操作 (Esc/Ctrl-C) | 0 | 无或 "Cancelled" |
| `opencc web` + Ctrl-C | 0 | "Web server stopped." |
| `opencc web stop` 但服务未运行 | 0 | "Web server is not running." |
| `opencc pick` 用户取消 | 0 | 无 |
| 配置/Provider 不存在 | 0 | "Configuration 'xxx' not found." |
| 无 provider 配置 | 0 | "No providers configured. Run 'opencc config' to set up." |
| Profile 不存在 | 0 | "Profile 'xxx' not found." |

以下情况应该 exit 1（真正的错误）：

| 场景 | Exit Code | 说明 |
|------|-----------|------|
| 配置文件损坏/无法解析 | 1 | JSON 语法错误等 |
| 网络/IO 错误 | 1 | 无法连接、文件权限等 |
| 子进程异常退出 | 传递子进程 exit code | claude 等 CLI 的退出码 |

---

## Open Questions

1. **Minimum terminal width**: 80 or 100 columns for split view?
2. **Fallback for narrow terminals**: Stack vertically or hide detail?
3. **Keyboard shortcuts**: Keep vim-style (hjkl) or standard arrows only?
4. **Color theme**: Keep current soft palette or refresh?

---

*Created: 2025-02-14*
*Status: Implementation in Progress*

## Implementation Status

- [x] Phase 1: Core Components
  - [x] `components/list.go` - Generic list with cursor
  - [x] `components/form.go` - Generic form with fields
  - [x] `components/keyvalue.go` - Key-value editor
- [x] Phase 2: Layout
  - [x] `menu.go` - Main menu
  - [x] `dashboard.go` - Split-view dashboard
  - [x] `settings.go` - Global settings
  - [x] `launch.go` - Launch wizard
- [x] Phase 3: Integration
  - [x] Connect existing editors to dashboard
  - [x] Add `--new` flag to `opencc config`
  - [x] Settings API and Web UI
- [ ] Phase 4: Migration
  - [ ] Make new TUI the default
  - [ ] Remove old TUI after testing

