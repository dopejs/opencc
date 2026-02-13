# OpenCC Roadmap

## Next: v1.5.0 - Multi-Provider Transform & CLI Support

### Feature 1: Provider Request Transform
Different providers (OpenAI, Azure, Bedrock, etc.) have different API formats.
Need to transform request/response bodies based on provider type.

Considerations:
- Provider config needs a `type` or `format` field
- Transform layer between proxy and upstream
- Handle model name mapping per provider
- Handle auth header differences

### Feature 2: Multi-CLI Support
Support launching different AI coding tools, not just Claude Code.

Target CLIs:
- Claude Code (current default)
- Codex
- OpenCode

Considerations:
- CLI selection in config or command flag
- Different CLIs may need different env vars
- Request transform depends on target CLI

### TUI Redesign Consideration
Current TUI is getting complex with:
- Providers (8+ fields + env vars)
- Profiles (providers + routing + threshold)
- Scenario routing (5 scenarios × providers × models)
- Project bindings

Options to consider:
1. Hierarchical navigation (tree-style)
2. Tab-based sections
3. Search/filter for large lists
4. Wizard-style for complex setup
5. Rely more on Web UI for complex config, keep TUI simple

---
*Created: 2026-02-13*
*Status: Planning*
