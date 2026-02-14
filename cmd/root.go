package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/internal/proxy"
	"github.com/dopejs/opencc/tui"
	"github.com/spf13/cobra"
)

// stdinReader is the reader used for interactive prompts. Tests can replace it.
var stdinReader io.Reader = os.Stdin

var Version = "1.5.3"

var rootCmd = &cobra.Command{
	Use:   "opencc [cli args...]",
	Short: "Multi-CLI environment switcher with proxy failover",
	Long:  "Load environment variables and start CLI (Claude Code, Codex, or OpenCode) with proxy failover.",
	// Allow unknown flags to pass through to claude
	DisableFlagParsing: false,
	SilenceUsage:       true,
	SilenceErrors:      true,
	RunE:               runProxy,
}

var cliFlag string
var legacyTUI bool

func init() {
	// -p/--profile is the new flag, -f/--fallback is kept for backward compatibility but hidden
	rootCmd.Flags().StringP("profile", "p", "", "profile name (use -p without value to pick interactively)")
	rootCmd.Flags().Lookup("profile").NoOptDefVal = " "
	rootCmd.Flags().StringP("fallback", "f", "", "alias for --profile (deprecated)")
	rootCmd.Flags().Lookup("fallback").NoOptDefVal = " "
	rootCmd.Flags().Lookup("fallback").Hidden = true
	rootCmd.Flags().StringVar(&cliFlag, "cli", "", "CLI to use (claude, codex, opencode)")
	rootCmd.Flags().BoolVar(&legacyTUI, "legacy", false, "use legacy TUI interface")
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(bindCmd)
	rootCmd.AddCommand(unbindCmd)
	rootCmd.AddCommand(statusCmd)

	// Set custom help function only for root command
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			cmd.Println(rootHelpText(cmd))
		} else {
			defaultHelp(cmd, args)
		}
	})
}

func rootHelpText(cmd *cobra.Command) string {
	return fmt.Sprintf(`%s

Usage:
  %s
  %s [command]

Quick Start:
  opencc                       Start with default profile
  opencc -p <profile>          Start with specific profile
  opencc --cli codex           Start with specific CLI
  opencc config                Open TUI configuration

Configuration:
  config                       Open TUI to manage providers and profiles
  config add provider [name]   Add a new provider
  config add profile [name]    Add a new profile
  config edit provider <name>  Edit an existing provider
  config delete provider <name> Delete a provider

Project Binding:
  bind <profile>               Bind current directory to a profile
  bind --cli <cli>             Bind current directory to a CLI
  unbind                       Remove binding for current directory
  status                       Show binding status

Web Interface:
  web                          Start web UI (foreground, opens browser)
  web -d                       Start web UI (background daemon)
  web stop                     Stop web daemon
  web status                   Show web daemon status
  web enable                   Install as system service (auto-start)
  web disable                  Uninstall system service

Other Commands:
  list                         List all providers and profiles
  pick                         Interactively select providers
  use <provider>               Use a specific provider directly
  upgrade                      Upgrade to latest version
  version                      Show version
  completion                   Generate shell completion

Flags:
%s
Use "%s [command] --help" for more information about a command.`,
		cmd.Long,
		cmd.UseLine(),
		cmd.CommandPath(),
		cmd.LocalFlags().FlagUsages(),
		cmd.CommandPath())
}

func Execute() error {
	// Pre-process: when -p/--profile or -f/--fallback uses NoOptDefVal, cobra won't consume
	// the next arg as its value. Merge "-p <name>" into "-p=<name>" so that
	// cobra parses it correctly and doesn't treat <name> as a subcommand.
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" || args[i] == "--profile" || args[i] == "-f" || args[i] == "--fallback" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				args[i] = args[i] + "=" + args[i+1]
				args = append(args[:i+1], args[i+2:]...)
			}
			break
		}
		// Stop if we hit a non-flag arg (subcommand) before -p/-f
		if !strings.HasPrefix(args[i], "-") {
			break
		}
	}
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func runProxy(cmd *cobra.Command, args []string) error {
	// Support both -p/--profile (new) and -f/--fallback (deprecated)
	profileFlag, _ := cmd.Flags().GetString("profile")
	if profileFlag == "" {
		profileFlag, _ = cmd.Flags().GetString("fallback")
	}

	providerNames, profile, cli, err := resolveProviderNamesAndCLI(profileFlag, cliFlag)
	if err != nil {
		return err
	}

	providerNames, err = validateProviderNames(providerNames, profile)
	if err != nil {
		return err
	}

	// Get the full profile config for routing support
	pc := config.GetProfileConfig(profile)

	return startProxy(providerNames, pc, cli, args)
}

func startProxy(names []string, pc *config.ProfileConfig, cli string, args []string) error {
	providers, err := buildProviders(names)
	if err != nil {
		return err
	}

	// Set up logger
	logDir := config.ConfigDirPath()
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "proxy.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile = nil
	}

	var logger *log.Logger
	if logFile != nil {
		logger = log.New(logFile, "", log.LstdFlags)
		defer logFile.Close()
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	// Initialize structured logger for web API access
	if err := proxy.InitGlobalLogger(logDir); err != nil {
		logger.Printf("Warning: failed to initialize structured logger: %v", err)
	}

	logger.Printf("Starting proxy with %d providers:", len(providers))
	for i, p := range providers {
		logger.Printf("  [%d] %s → %s (model=%s)", i+1, p.Name, p.BaseURL.String(), p.Model)
	}

	// Use CLI from parameter (already resolved from flag/binding/default)
	cliBin := cli
	if cliBin == "" {
		cliBin = "claude"
	}

	// Determine client format based on CLI type
	clientFormat := GetCLIClientFormat(GetCLIType(cliBin))
	logger.Printf("CLI: %s, Client format: %s", cliBin, clientFormat)

	// Start proxy — with routing if configured, otherwise plain
	var port int
	if pc != nil && len(pc.Routing) > 0 {
		routingCfg, err := buildRoutingConfig(pc, providers, logger)
		if err != nil {
			return fmt.Errorf("failed to build routing config: %w", err)
		}
		port, err = proxy.StartProxyWithRouting(routingCfg, clientFormat, "127.0.0.1:0", logger)
		if err != nil {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
	} else {
		port, err = proxy.StartProxy(providers, clientFormat, "127.0.0.1:0", logger)
		if err != nil {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
	}

	logger.Printf("Proxy listening on 127.0.0.1:%d", port)

	// Merge env_vars from all providers for this specific CLI
	// For numeric values like ANTHROPIC_MAX_CONTEXT_WINDOW, use the minimum value
	// This ensures the CLI respects the most restrictive provider's limit
	mergedEnvVars := mergeProviderEnvVarsForCLI(providers, cliBin)
	for k, v := range mergedEnvVars {
		os.Setenv(k, v)
		logger.Printf("Setting env: %s=%s", k, v)
	}

	// Set environment variables based on CLI type
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	setupCLIEnvironment(cliBin, proxyURL, logger)

	// Find CLI binary
	cliPath, err := exec.LookPath(cliBin)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", cliBin, err)
	}

	// Start CLI as subprocess (not exec, so proxy stays alive)
	cliCmd := exec.Command(cliPath, args...)
	cliCmd.Stdin = os.Stdin
	cliCmd.Stdout = os.Stdout
	cliCmd.Stderr = os.Stderr

	// Forward signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if cliCmd.Process != nil {
				cliCmd.Process.Signal(sig)
			}
		}
	}()

	if err := cliCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func buildProviders(names []string) ([]*proxy.Provider, error) {
	var providers []*proxy.Provider

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		p := config.GetProvider(name)
		if p == nil {
			return nil, fmt.Errorf("configuration '%s' not found", name)
		}

		if p.BaseURL == "" || p.AuthToken == "" {
			return nil, fmt.Errorf("%s missing base_url or auth_token", name)
		}

		model := p.Model
		if model == "" {
			model = "claude-sonnet-4-5"
		}
		reasoningModel := p.ReasoningModel
		if reasoningModel == "" {
			reasoningModel = "claude-sonnet-4-5-thinking"
		}
		haikuModel := p.HaikuModel
		if haikuModel == "" {
			haikuModel = "claude-haiku-4-5"
		}
		opusModel := p.OpusModel
		if opusModel == "" {
			opusModel = "claude-opus-4-5"
		}
		sonnetModel := p.SonnetModel
		if sonnetModel == "" {
			sonnetModel = "claude-sonnet-4-5"
		}

		u, err := url.Parse(p.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL for provider %s: %w", name, err)
		}

		providers = append(providers, &proxy.Provider{
			Name:            name,
			Type:            p.GetType(),
			BaseURL:         u,
			Token:           p.AuthToken,
			Model:           model,
			ReasoningModel:  reasoningModel,
			HaikuModel:      haikuModel,
			OpusModel:       opusModel,
			SonnetModel:     sonnetModel,
			EnvVars:         p.EnvVars,
			ClaudeEnvVars:   p.ClaudeEnvVars,
			CodexEnvVars:    p.CodexEnvVars,
			OpenCodeEnvVars: p.OpenCodeEnvVars,
			Healthy:         true,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers")
	}
	return providers, nil
}

// mergeProviderEnvVarsForCLI merges env_vars from all providers for a specific CLI.
// For numeric values like ANTHROPIC_MAX_CONTEXT_WINDOW, uses the minimum value.
// For other values, first provider's value takes precedence.
func mergeProviderEnvVarsForCLI(providers []*proxy.Provider, cli string) map[string]string {
	result := make(map[string]string)

	// Env vars where we should take the minimum numeric value
	minValueKeys := map[string]bool{
		"ANTHROPIC_MAX_CONTEXT_WINDOW":          true,
		"OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": true,
	}

	for _, p := range providers {
		envVars := p.GetEnvVarsForCLI(cli)
		if envVars == nil {
			continue
		}
		for k, v := range envVars {
			if k == "" || v == "" {
				continue
			}

			existing, exists := result[k]
			if !exists {
				result[k] = v
				continue
			}

			// For min-value keys, compare and keep the smaller value
			if minValueKeys[k] {
				existingVal, err1 := strconv.Atoi(existing)
				newVal, err2 := strconv.Atoi(v)
				if err1 == nil && err2 == nil && newVal < existingVal {
					result[k] = v
				}
			}
			// For other keys, first value wins (already set)
		}
	}

	return result
}

// buildRoutingConfig creates a RoutingConfig from a ProfileConfig.
// Provider instances are shared across scenarios: same name → same *Provider pointer.
func buildRoutingConfig(pc *config.ProfileConfig, defaultProviders []*proxy.Provider, logger *log.Logger) (*proxy.RoutingConfig, error) {
	// Build a map of all provider instances by name (from default providers)
	providerMap := make(map[string]*proxy.Provider)
	for _, p := range defaultProviders {
		providerMap[p.Name] = p
	}

	// Also build providers for any names that only appear in routing scenarios
	for _, route := range pc.Routing {
		for _, pr := range route.Providers {
			if _, ok := providerMap[pr.Name]; !ok {
				// Need to build this provider
				ps, err := buildProviders([]string{pr.Name})
				if err != nil {
					logger.Printf("[routing] skipping unknown provider %q in routing: %v", pr.Name, err)
					continue
				}
				providerMap[pr.Name] = ps[0]
			}
		}
	}

	// Build scenario routes
	scenarioRoutes := make(map[config.Scenario]*proxy.ScenarioProviders)
	for scenario, route := range pc.Routing {
		var chain []*proxy.Provider
		models := make(map[string]string)
		for _, pr := range route.Providers {
			if p, ok := providerMap[pr.Name]; ok {
				chain = append(chain, p)
				if pr.Model != "" {
					models[pr.Name] = pr.Model
				}
			}
		}
		if len(chain) > 0 {
			scenarioRoutes[scenario] = &proxy.ScenarioProviders{
				Providers: chain,
				Models:    models,
			}
			logger.Printf("[routing] scenario %s: %d providers, %d model overrides", scenario, len(chain), len(models))
		}
	}

	return &proxy.RoutingConfig{
		DefaultProviders:     defaultProviders,
		ScenarioRoutes:       scenarioRoutes,
		LongContextThreshold: pc.LongContextThreshold,
	}, nil
}

// resolveProviderNamesAndCLI determines the provider list and CLI based on flags and bindings.
// Returns the provider names, the profile used, and the CLI to use.
func resolveProviderNamesAndCLI(profileFlag string, cliFlag string) ([]string, string, string, error) {
	// Determine CLI: flag > binding > default
	cli := cliFlag

	// -f (no value, NoOptDefVal=" ") → interactive profile picker
	if profileFlag == " " {
		profile, err := tui.RunProfilePicker()
		if err != nil {
			return nil, "", "", err
		}
		names, err := config.ReadProfileOrder(profile)
		if err != nil {
			return nil, "", "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		if len(names) == 0 {
			return nil, "", "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		if cli == "" {
			cli = config.GetDefaultCLI()
		}
		return names, profile, cli, nil
	}

	// -f <name> → use that specific profile
	if profileFlag != "" {
		names, err := config.ReadProfileOrder(profileFlag)
		if err != nil {
			return nil, "", "", fmt.Errorf("profile '%s' not found", profileFlag)
		}
		if len(names) == 0 {
			return nil, "", "", fmt.Errorf("profile '%s' has no providers configured", profileFlag)
		}
		if cli == "" {
			cli = config.GetDefaultCLI()
		}
		return names, profileFlag, cli, nil
	}

	// No profile flag → check for project binding first
	cwd, err := os.Getwd()
	if err == nil {
		cwd = filepath.Clean(cwd)
		if binding := config.GetProjectBinding(cwd); binding != nil {
			// Found project binding
			profile := binding.Profile
			if profile == "" {
				profile = config.GetDefaultProfile()
			}

			// Use binding CLI if not overridden by flag
			if cli == "" && binding.CLI != "" {
				cli = binding.CLI
			}

			names, err := config.ReadProfileOrder(profile)
			if err == nil && len(names) > 0 {
				if cli == "" {
					cli = config.GetDefaultCLI()
				}
				return names, profile, cli, nil
			}
			// Profile was deleted, fall through to default
			if binding.Profile != "" {
				fmt.Fprintf(os.Stderr, "Warning: Bound profile '%s' not found, using default\n", binding.Profile)
			}
		}
	}

	// No binding → use default profile
	defaultProfile := config.GetDefaultProfile()
	fbNames, err := config.ReadFallbackOrder()
	if err == nil && len(fbNames) > 0 {
		if cli == "" {
			cli = config.GetDefaultCLI()
		}
		return fbNames, defaultProfile, cli, nil
	}

	// default profile missing or empty — interactive selection
	names, err := interactiveSelectProviders()
	if err != nil {
		return nil, "", "", err
	}
	if names == nil {
		// User cancelled
		return nil, "", "", fmt.Errorf("cancelled")
	}
	if cli == "" {
		cli = config.GetDefaultCLI()
	}
	return names, defaultProfile, cli, nil
}

// interactiveSelectProviders uses TUI to select providers.
// If no providers exist, launches the create-first editor.
// Otherwise launches the checkbox picker.
// Returns nil, nil if user cancels.
func interactiveSelectProviders() ([]string, error) {
	available := config.ProviderNames()
	if len(available) == 0 {
		// No providers at all — launch TUI editor to create one
		name, err := tui.RunCreateFirst()
		if err != nil {
			// User cancelled
			return nil, nil
		}
		if name == "" {
			return nil, nil
		}
		return []string{name}, nil
	}

	// Providers exist but no default profile — launch picker
	selected, err := tui.RunPick()
	if err != nil {
		// User cancelled
		return nil, nil
	}
	if len(selected) == 0 {
		return nil, nil
	}

	// Write selection to default profile
	if err := config.WriteFallbackOrder(selected); err != nil {
		return nil, fmt.Errorf("failed to save fallback order: %w", err)
	}
	fmt.Printf("Saved fallback order: %s\n", strings.Join(selected, ", "))

	return selected, nil
}

// validateProviderNames checks that each provider exists in the config.
// Prompts user to confirm removal of missing providers from the profile.
func validateProviderNames(names []string, profile string) ([]string, error) {
	var valid, missing []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if config.GetProvider(name) == nil {
			missing = append(missing, name)
		} else {
			valid = append(valid, name)
		}
	}

	if len(missing) == 0 {
		return names, nil
	}

	fmt.Printf("%s provider(s) not found. Continue and remove from profile? (y/n): ", strings.Join(missing, ", "))
	reader := bufio.NewReader(stdinReader)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	if answer != "y" && answer != "yes" {
		return nil, fmt.Errorf("aborted")
	}

	// Remove missing from profile
	for _, name := range missing {
		config.RemoveFromProfileOrder(profile, name)
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid providers remaining. Run 'opencc config' to set up providers")
	}

	return valid, nil
}

// CLIType represents the type of CLI being used.
type CLIType string

const (
	CLIClaude   CLIType = "claude"
	CLICodex    CLIType = "codex"
	CLIOpenCode CLIType = "opencode"
)

// GetCLIType returns the CLI type from the binary name.
func GetCLIType(cliBin string) CLIType {
	switch cliBin {
	case "codex":
		return CLICodex
	case "opencode":
		return CLIOpenCode
	default:
		return CLIClaude
	}
}

// GetCLIClientFormat returns the API format used by the CLI.
func GetCLIClientFormat(cliType CLIType) string {
	switch cliType {
	case CLICodex:
		return config.ProviderTypeOpenAI
	default:
		// Claude Code and OpenCode use Anthropic format by default
		return config.ProviderTypeAnthropic
	}
}

// setupCLIEnvironment sets the appropriate environment variables for the CLI.
func setupCLIEnvironment(cliBin string, proxyURL string, logger *log.Logger) {
	cliType := GetCLIType(cliBin)

	switch cliType {
	case CLICodex:
		// Codex uses OpenAI environment variables
		os.Setenv("OPENAI_BASE_URL", proxyURL)
		os.Setenv("OPENAI_API_KEY", "opencc-proxy")
		logger.Printf("Setting Codex env: OPENAI_BASE_URL=%s", proxyURL)

	case CLIOpenCode:
		// OpenCode supports multiple providers, set both
		// It will use the appropriate one based on the model prefix
		os.Setenv("ANTHROPIC_BASE_URL", proxyURL)
		os.Setenv("ANTHROPIC_API_KEY", "opencc-proxy")
		os.Setenv("OPENAI_BASE_URL", proxyURL)
		os.Setenv("OPENAI_API_KEY", "opencc-proxy")
		logger.Printf("Setting OpenCode env: ANTHROPIC_BASE_URL=%s, OPENAI_BASE_URL=%s", proxyURL, proxyURL)

	default:
		// Claude Code uses Anthropic environment variables
		os.Setenv("ANTHROPIC_BASE_URL", proxyURL)
		os.Setenv("ANTHROPIC_AUTH_TOKEN", "opencc-proxy")
		logger.Printf("Setting Claude env: ANTHROPIC_BASE_URL=%s", proxyURL)
	}
}

