package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/dopejs/opencc/internal/config"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:               "use <config> [claude args...]",
	Short:             "Load config and exec claude directly",
	ValidArgsFunction: completeConfigNames,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE:              runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	available := config.ProviderNames()

	if len(args) == 0 {
		fmt.Println("Usage: opencc use <provider> [claude args...]")
		if len(available) > 0 {
			fmt.Printf("\nAvailable providers: %s\n", strings.Join(available, ", "))
		} else {
			fmt.Println("\nNo providers configured. Run 'opencc config' to set up providers.")
		}
		return nil
	}

	configName := args[0]
	claudeArgs := args[1:]

	if err := config.ExportProviderToEnv(configName); err != nil {
		fmt.Printf("Provider '%s' not found.\n", configName)
		if len(available) > 0 {
			fmt.Printf("Available providers: %s\n", strings.Join(available, ", "))
		} else {
			fmt.Println("No providers configured. Run 'opencc config' to set up providers.")
		}
		return nil
	}

	// Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	// Replace process with claude (like shell exec)
	argv := append([]string{"claude"}, claudeArgs...)
	return syscall.Exec(claudeBin, argv, os.Environ())
}

func completeConfigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	names := config.ProviderNames()
	return names, cobra.ShellCompDirectiveNoFileComp
}
