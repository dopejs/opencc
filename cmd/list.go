package cmd

import (
	"github.com/dopejs/opencc/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List providers and groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunDetailList(tui.ListViewAll)
	},
}

var listProviderCmd = &cobra.Command{
	Use:   "provider",
	Short: "List providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunDetailList(tui.ListViewProviders)
	},
}

var listGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "List groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunDetailList(tui.ListViewGroups)
	},
}

func init() {
	listCmd.AddCommand(listProviderCmd)
	listCmd.AddCommand(listGroupCmd)
}
