package cmd

import (
	"fmt"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/tui"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:           "pick [claude args...]",
	Short:         "Select providers interactively and start proxy",
	Long:          "Launch a checkbox picker to select providers for this session, then start the proxy.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runPick,
}

func runPick(cmd *cobra.Command, args []string) error {
	available := config.ProviderNames()
	if len(available) == 0 {
		return fmt.Errorf("no providers configured. Run 'opencc config' to set up providers")
	}

	selected, err := tui.RunPick()
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no providers selected")
	}

	return startProxy(selected, args)
}
