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
cmd/           # Cobra commands (root, config, pick, web, upgrade, bind)
internal/
  config/      # Config store, types, legacy migration, compat helpers
  daemon/      # Web server daemon management (start/stop, platform-specific)
  proxy/       # Reverse proxy with failover, token calculation, session cache
  web/         # HTTP API server + embedded static frontend
    static/    # app.js, index.html, style.css (vanilla JS, no build step)
tui/           # All TUI models (editor, pick, config_main, fallback, routing, etc.)
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

- **Commit often**: Each completed task/item should be committed individually, not batched into one large commit. After finishing a feature or fix, commit immediately before moving to the next task.
- **Pre-release check**: Before tagging a release, check for unpushed commits (`git log origin/main..HEAD`) and push them first.
- **Update version constant**: Before releasing, update `Version` in `cmd/root.go` to match the release tag.
- **Update README**: Before releasing, check that `README.md` reflects all new features and changes.
- **No summary/explanation docs**: Do NOT create markdown files to summarize or explain completed work. No "implementation notes", no "changes summary", no "feature documentation". The commit message and code comments are sufficient.
- **Keep planning docs**: Architecture planning and design docs should be kept for context across sessions. Store them in a `docs/` folder if needed.
- **No example files**: Do NOT create example config files (*.json, *.yaml, etc.) in the repository root. Examples belong in README.md or `docs/`.
- **Minimal test files**: Only add tests for new public APIs or complex logic. Do not create excessive test files for simple functions. Prefer table-driven tests in existing *_test.go files.
- **No unnecessary files**: Before committing, review `git status` and remove any generated, temporary, or example files that should not be in the repository.

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
- v1.3.1: Download progress bar, README refresh
- v1.3.2: Fix progress bar display (show downloaded/total size)
- v1.4.0: Scenario routing, token calculation, session cache, project bindings
