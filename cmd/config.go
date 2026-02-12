package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dopejs/opencc/internal/config"
	"github.com/dopejs/opencc/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage providers and profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := tui.RunConfigMain()
		if err != nil && err.Error() == "cancelled" {
			return nil
		}
		return err
	},
}

// --- add subcommands ---

var configAddCmd = &cobra.Command{
	Use:   "add [provider|profile]",
	Short: "Add a provider or profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		typ, err := tui.RunSelectType()
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		switch typ {
		case "provider":
			_, err := tui.RunAddProvider("")
			if err != nil && err.Error() == "cancelled" {
				return nil
			}
			return err
		case "group":
			err := tui.RunAddGroup("")
			if err != nil && err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		return nil
	},
}

var configAddProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Add a new provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		_, err := tui.RunAddProvider(name)
		if err != nil && err.Error() == "cancelled" {
			return nil
		}
		return err
	},
}

var configAddGroupCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Add a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		err := tui.RunAddGroup(name)
		if err != nil && err.Error() == "cancelled" {
			return nil
		}
		return err
	},
}

// --- delete subcommands ---

var configDeleteCmd = &cobra.Command{
	Use:   "delete [provider|profile] [name]",
	Short: "Delete a provider or profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		typ, err := tui.RunSelectType()
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		switch typ {
		case "provider":
			return deleteProvider("")
		case "group":
			return deleteGroup("")
		}
		return nil
	},
}

var configDeleteProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Delete a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return deleteProvider(name)
	},
}

var configDeleteGroupCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Delete a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return deleteGroup(name)
	},
}

func deleteProvider(name string) error {
	names := config.ProviderNames()
	if len(names) == 0 {
		fmt.Println("No providers configured.")
		return nil
	}
	if len(names) == 1 {
		fmt.Println("Cannot delete the last provider. At least one provider must remain.")
		return nil
	}

	if name == "" {
		var err error
		name, err = tui.RunSelectProvider(1)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
	}

	if config.GetProvider(name) == nil {
		fmt.Printf("Provider %q not found.\n", name)
		return nil
	}

	if !confirmDelete(name) {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.DeleteProviderByName(name); err != nil {
		return err
	}
	fmt.Printf("Deleted provider %q.\n", name)
	return nil
}

func deleteGroup(name string) error {
	if name == "" {
		var err error
		name, err = tui.RunSelectGroup(true) // excludeDefault=true
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
	}

	if name == "default" {
		fmt.Println("Cannot delete the default profile.")
		return nil
	}

	profiles := config.ListProfiles()
	found := false
	for _, p := range profiles {
		if p == name {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Profile %q not found.\n", name)
		return nil
	}

	if !confirmDelete(name) {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.DeleteProfile(name); err != nil {
		return err
	}
	fmt.Printf("Deleted profile %q.\n", name)
	return nil
}

func confirmDelete(name string) bool {
	fmt.Printf("Delete '%s'? This cannot be undone. (y/n): ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// --- edit subcommands ---

var configEditCmd = &cobra.Command{
	Use:   "edit [provider|profile] [name]",
	Short: "Edit a provider or profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		typ, err := tui.RunSelectType()
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
		switch typ {
		case "provider":
			return editProvider("")
		case "group":
			return editGroup("")
		}
		return nil
	},
}

var configEditProviderCmd = &cobra.Command{
	Use:   "provider [name]",
	Short: "Edit a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return editProvider(name)
	},
}

var configEditGroupCmd = &cobra.Command{
	Use:   "profile [name]",
	Short: "Edit a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return editGroup(name)
	},
}

func editProvider(name string) error {
	if name == "" {
		var err error
		name, err = tui.RunSelectProvider(0)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
	}

	if config.GetProvider(name) == nil {
		fmt.Printf("Provider %q not found.\n", name)
		return nil
	}

	err := tui.RunEditProvider(name)
	if err != nil && err.Error() == "cancelled" {
		return nil
	}
	return err
}

func editGroup(name string) error {
	if name == "" {
		var err error
		name, err = tui.RunSelectGroup(false)
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}
	}

	profiles := config.ListProfiles()
	found := false
	for _, p := range profiles {
		if p == name {
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Profile %q not found.\n", name)
		return nil
	}

	err := tui.RunEditGroup(name)
	if err != nil && err.Error() == "cancelled" {
		return nil
	}
	return err
}

func init() {
	configAddCmd.AddCommand(configAddProviderCmd)
	configAddCmd.AddCommand(configAddGroupCmd)

	configDeleteCmd.AddCommand(configDeleteProviderCmd)
	configDeleteCmd.AddCommand(configDeleteGroupCmd)

	configEditCmd.AddCommand(configEditProviderCmd)
	configEditCmd.AddCommand(configEditGroupCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configEditCmd)
}
