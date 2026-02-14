# opencc

[简体中文](docs/README.zh-CN.md) | [繁體中文](docs/README.zh-TW.md) | [Español](docs/README.es.md)

Multi-CLI environment switcher for Claude Code, Codex, and OpenCode with API proxy auto-failover.

## Features

- **Multi-CLI Support** — Supports Claude Code, Codex, and OpenCode, configurable per project
- **Multi-Config Management** — Manage all API configurations in `~/.opencc/opencc.json`
- **Proxy Failover** — Built-in HTTP proxy that automatically switches to backup providers when the primary is unavailable
- **Scenario Routing** — Intelligent routing based on request characteristics (thinking, image, longContext, etc.)
- **Project Bindings** — Bind directories to specific profiles and CLIs for project-level auto-configuration
- **Environment Variables** — Configure CLI-specific environment variables at the provider level
- **TUI Config Interface** — Interactive terminal UI with Dashboard and legacy modes
- **Web Management UI** — Browser-based visual management for providers, profiles, and project bindings
- **Self-Update** — One-command upgrade via `opencc upgrade` with semver version matching
- **Shell Completion** — Supports zsh / bash / fish

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh
```

Uninstall:

```sh
curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall
```

## Quick Start

```sh
# Open the TUI config interface and create your first provider
opencc config

# Launch (using default profile)
opencc

# Use a specific profile
opencc -p work

# Use a specific CLI
opencc --cli codex
```

## Command Reference

| Command | Description |
|---------|-------------|
| `opencc` | Launch CLI (using project binding or default config) |
| `opencc -p <profile>` | Launch with a specific profile |
| `opencc -p` | Interactively select a profile |
| `opencc --cli <cli>` | Use a specific CLI (claude/codex/opencode) |
| `opencc use <provider>` | Directly use a specific provider (no proxy) |
| `opencc pick` | Interactively select a provider to launch |
| `opencc list` | List all providers and profiles |
| `opencc config` | Open the TUI config interface |
| `opencc config --legacy` | Use the legacy TUI interface |
| `opencc bind <profile>` | Bind current directory to a profile |
| `opencc bind --cli <cli>` | Bind current directory to a specific CLI |
| `opencc unbind` | Remove binding for current directory |
| `opencc status` | Show binding status for current directory |
| `opencc web start` | Start the Web management UI |
| `opencc web open` | Open the Web UI in browser |
| `opencc web stop` | Stop the Web server |
| `opencc upgrade` | Upgrade to the latest version |
| `opencc version` | Show version |

## Multi-CLI Support

opencc supports three AI coding assistant CLIs:

| CLI | Description | API Format |
|-----|-------------|------------|
| `claude` | Claude Code (default) | Anthropic Messages API |
| `codex` | OpenAI Codex CLI | OpenAI Chat Completions API |
| `opencode` | OpenCode | Anthropic / OpenAI |

### Set Default CLI

```sh
# Via TUI
opencc config  # Settings → Default CLI

# Via Web UI
opencc web open  # Settings page
```

### Per-Project CLI

```sh
cd ~/work/project
opencc bind --cli codex  # Use Codex for this directory
```

### Temporary CLI Override

```sh
opencc --cli opencode  # Use OpenCode for this session
```

## Profile Management

A profile is an ordered list of providers used for failover.

### Configuration Example

```json
{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}
```

### Using Profiles

```sh
# Use default profile
opencc

# Use a specific profile
opencc -p work

# Interactive selection
opencc -p
```

## Project Bindings

Bind directories to specific profiles and/or CLIs for project-level auto-configuration.

```sh
cd ~/work/company-project

# Bind profile
opencc bind work-profile

# Bind CLI
opencc bind --cli codex

# Bind both
opencc bind work-profile --cli codex

# Check status
opencc status

# Remove binding
opencc unbind
```

**Priority**: Command-line args > Project binding > Global default

## TUI Config Interface

```sh
opencc config
```

v1.5 introduces a new Dashboard interface:

- **Left panel**: Providers, Profiles, Project Bindings
- **Right panel**: Details for the selected item
- **Keyboard shortcuts**:
  - `a` - Add new item
  - `e` - Edit selected item
  - `d` - Delete selected item
  - `Tab` - Switch focus
  - `q` - Back / Quit

Use `--legacy` to switch to the legacy interface.

## Web Management UI

```sh
# Start (runs in background, port 19840)
opencc web start

# Open in browser
opencc web open

# Stop
opencc web stop
```

Web UI features:
- Provider and Profile management
- Project binding management
- Global settings (default CLI, default profile, port)
- Request log viewer
- Model field autocomplete

## Environment Variables

Each provider can have CLI-specific environment variables:

```json
{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}
```

### Common Claude Code Environment Variables

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_MAX_OUTPUT_TOKENS` | Max output tokens |
| `MAX_THINKING_TOKENS` | Extended thinking budget |
| `ANTHROPIC_MAX_CONTEXT_WINDOW` | Max context window |
| `BASH_DEFAULT_TIMEOUT_MS` | Bash default timeout |

## Scenario Routing

Automatically route requests to different providers based on request characteristics:

| Scenario | Trigger Condition |
|----------|-------------------|
| `think` | Thinking mode enabled |
| `image` | Contains image content |
| `longContext` | Content exceeds threshold |
| `webSearch` | Uses web_search tool |
| `background` | Uses Haiku model |

**Fallback mechanism**: If all providers in a scenario config fail, it automatically falls back to the profile's default providers.

Configuration example:

```json
{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}
```

## Config Files

| File | Description |
|------|-------------|
| `~/.opencc/opencc.json` | Main configuration file |
| `~/.opencc/proxy.log` | Proxy log |
| `~/.opencc/web.log` | Web server log |

### Full Configuration Example

```json
{
  "version": 5,
  "default_profile": "default",
  "default_cli": "claude",
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "cli": "codex"
    }
  }
}
```

## Upgrade

```sh
# Latest version
opencc upgrade

# Specific version
opencc upgrade 1.5
opencc upgrade 1.5.0
```

## Migrating from Older Versions

If you previously used the `~/.cc_envs/` format, opencc will automatically migrate to `~/.opencc/opencc.json`.

## Development

```sh
# Build
go build -o opencc .

# Test
go test ./...
```

Release: Push a tag and GitHub Actions will build automatically.

```sh
git tag v1.5.1
git push origin v1.5.1
```

## License

MIT
