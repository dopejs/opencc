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
	"strings"
	"syscall"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/internal/proxy"
	"github.com/dopejs/opencc/tui"
	"github.com/spf13/cobra"
)

// stdinReader is the reader used for interactive prompts. Tests can replace it.
var stdinReader io.Reader = os.Stdin

var Version = "1.2.0"

var rootCmd = &cobra.Command{
	Use:   "opencc [claude args...]",
	Short: "Claude Code environment switcher with fallback proxy",
	Long:  "Load environment variables and start Claude Code, optionally with a fallback proxy.",
	// Allow unknown flags to pass through to claude
	DisableFlagParsing: false,
	SilenceUsage:       true,
	SilenceErrors:      true,
	RunE:               runProxy,
}

func init() {
	rootCmd.Flags().StringP("fallback", "f", "", "fallback profile name (use -f without value to pick interactively)")
	rootCmd.Flags().Lookup("fallback").NoOptDefVal = " "
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(webCmd)
}

func Execute() error {
	// Pre-process: when -f/--fallback uses NoOptDefVal, cobra won't consume
	// the next arg as its value. Merge "-f <name>" into "-f=<name>" so that
	// cobra parses it correctly and doesn't treat <name> as a subcommand.
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" || args[i] == "--fallback" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				args[i] = args[i] + "=" + args[i+1]
				args = append(args[:i+1], args[i+2:]...)
			}
			break
		}
		// Stop if we hit a non-flag arg (subcommand) before -f
		if !strings.HasPrefix(args[i], "-") {
			break
		}
	}
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func runProxy(cmd *cobra.Command, args []string) error {
	profileFlag, _ := cmd.Flags().GetString("fallback")

	providerNames, profile, err := resolveProviderNames(profileFlag)
	if err != nil {
		return err
	}

	providerNames, err = validateProviderNames(providerNames, profile)
	if err != nil {
		return err
	}

	return startProxy(providerNames, args)
}

func startProxy(names []string, args []string) error {
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

	logger.Printf("Starting proxy with %d providers:", len(providers))
	for i, p := range providers {
		logger.Printf("  [%d] %s → %s (model=%s)", i+1, p.Name, p.BaseURL.String(), p.Model)
	}

	// Start embedded proxy
	port, err := proxy.StartProxy(providers, "127.0.0.1:0", logger)
	if err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	logger.Printf("Proxy listening on 127.0.0.1:%d", port)

	// Set environment for claude — proxy handles model mapping per-provider
	os.Setenv("ANTHROPIC_BASE_URL", fmt.Sprintf("http://127.0.0.1:%d", port))
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "opencc-proxy")

	// Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	// Start claude as subprocess (not exec, so proxy stays alive)
	claudeCmd := exec.Command(claudeBin, args...)
	claudeCmd.Stdin = os.Stdin
	claudeCmd.Stdout = os.Stdout
	claudeCmd.Stderr = os.Stderr

	// Forward signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if claudeCmd.Process != nil {
				claudeCmd.Process.Signal(sig)
			}
		}
	}()

	if err := claudeCmd.Run(); err != nil {
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
			Name:           name,
			BaseURL:        u,
			Token:          p.AuthToken,
			Model:          model,
			ReasoningModel: reasoningModel,
			HaikuModel:     haikuModel,
			OpusModel:      opusModel,
			SonnetModel:    sonnetModel,
			Healthy:        true,
		})
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no valid providers")
	}
	return providers, nil
}

// resolveProviderNames determines the provider list based on the -f flag value.
// Returns the provider names and the profile used.
func resolveProviderNames(profileFlag string) ([]string, string, error) {
	// -f (no value, NoOptDefVal=" ") → interactive profile picker
	if profileFlag == " " {
		profile, err := tui.RunProfilePicker()
		if err != nil {
			return nil, "", err
		}
		names, err := config.ReadProfileOrder(profile)
		if err != nil {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		if len(names) == 0 {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profile)
		}
		return names, profile, nil
	}

	// -f <name> → use that specific profile
	if profileFlag != "" {
		names, err := config.ReadProfileOrder(profileFlag)
		if err != nil {
			return nil, "", fmt.Errorf("profile '%s' not found", profileFlag)
		}
		if len(names) == 0 {
			return nil, "", fmt.Errorf("profile '%s' has no providers configured", profileFlag)
		}
		return names, profileFlag, nil
	}

	// No flag → existing behavior (default profile, or interactive selection)
	fbNames, err := config.ReadFallbackOrder()
	if err == nil && len(fbNames) > 0 {
		return fbNames, "default", nil
	}

	// default profile missing or empty — interactive selection
	names, err := interactiveSelectProviders()
	if err != nil {
		return nil, "", err
	}
	return names, "default", nil
}

// interactiveSelectProviders uses TUI to select providers.
// If no providers exist, launches the create-first editor.
// Otherwise launches the checkbox picker.
func interactiveSelectProviders() ([]string, error) {
	available := config.ProviderNames()
	if len(available) == 0 {
		// No providers at all — launch TUI editor to create one
		name, err := tui.RunCreateFirst()
		if err != nil {
			return nil, fmt.Errorf("no providers configured")
		}
		if name == "" {
			return nil, fmt.Errorf("no providers configured")
		}
		return []string{name}, nil
	}

	// Providers exist but no default profile — launch picker
	selected, err := tui.RunPick()
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no providers selected")
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
