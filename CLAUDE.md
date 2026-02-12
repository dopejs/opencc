# OpenCC - Claude Code Environment Switcher

## Project Overview

Go CLI tool for managing multiple Claude API provider configurations with proxy failover and TUI/Web interfaces.

## Tech Stack

- **Language**: Go
- **TUI**: Bubble Tea (charmbracelet/bubbletea) + Lip Gloss
- **Web**: Embedded static files (HTML/CSS/JS, no framework), Go HTTP server
- **CLI**: Cobra
- **Config**: JSON store at `~/.opencc/opencc.json`

## Project Structure

```
cmd/           # Cobra commands (root, config, pick, web, upgrade)
internal/
  config/      # Config store, types, legacy migration, compat helpers
  daemon/      # Web server daemon management (start/stop, platform-specific)
  proxy/       # Reverse proxy with failover
  web/         # HTTP API server + embedded static frontend
    static/    # app.js, index.html, style.css (vanilla JS, no build step)
tui/           # All TUI models (editor, pick, config_main, fallback, etc.)
```

## Build & Test

```sh
go build ./...
go test ./...
```

## Release Process

Push a git tag to trigger GitHub Actions release workflow:

```sh
git tag v1.x.0
git push origin v1.x.0
```

Do NOT use `gh release create` — the CI pipeline handles release creation automatically.

## Workflow Rules

- **Commit often**: Each completed task/item should be committed individually, not batched into one large commit.
- **Pre-release check**: Before tagging a release, check for unpushed commits (`git log origin/main..HEAD`) and push them first.

## Key Conventions

- Config convenience functions in `internal/config/compat.go` wrap `DefaultStore()` methods
- TUI models follow Bubble Tea pattern: `newXxxModel()`, `Init()`, `Update()`, `View()`
- Standalone TUI entry points: `RunXxx()` functions in tui package
- Inline config sub-editors use wrapper types implementing `tea.Model` (e.g. `editorWrapper`, `fallbackWrapper`)
- Web API routes: `/api/v1/providers`, `/api/v1/profiles`, `/api/v1/health`, `/api/v1/reload`
- Web frontend uses vanilla JS (no build tools), CSS custom properties for theming
- Model IDs in autocomplete must be verified against official Anthropic docs

## Version History

- v1.0.0: Initial release
- v1.1.0: Go rewrite with proxy failover and TUI
- v1.1.1: Upgrade command, sorted configs
- v1.2.0: Fallback profiles, pick command, installer
- v1.3.0: Web UI, profile assignment on provider add, model autocomplete, CLI name args
