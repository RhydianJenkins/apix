package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/handlers"
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags "-X main.version=...".
// See the `build`/`release` apps and the `ldflags` in flake.nix.
var version = "dev"

func main() {
	if err := newRootCmd(version).Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd(version string) *cobra.Command {
	var cfgPathFlag string

	rootCmd := &cobra.Command{
		Use:   "apix [command]",
		Short: "API eXecuter (APIX) is a CLI tool to manage API domains and make requests",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cfgPathFlag == "" {
				return nil
			}

			expanded, err := expandConfigPath(cfgPathFlag)
			if err != nil {
				return err
			}

			config.CfgPath = expanded

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgPathFlag, "config", "c", "", "path to the apix config file (default ~/.apix.yaml)")

	newCmd := &cobra.Command{
		Use:     "new [name] [base]",
		Short:   "Create a new API domain",
		Example: "apix new myapi https://api.example.com --user foo --pass bar",
		Args:    cobra.ExactArgs(2),
		Run:     handlers.NewHandler,
	}
	newCmd.Flags().String("protocol", config.ProtocolHTTP, "protocol for this domain: http or grpc")
	newCmd.Flags().StringSliceP("header", "H", []string{}, "default headers/metadata for this domain in format 'Key: Value' (can be used multiple times)")
	newCmd.Flags().Int("port", 0, "port to connect to; base must be host-only when set (required for grpc, optional for http)")
	addHTTPNewFlags(newCmd)
	addGRPCNewFlags(newCmd)
	rootCmd.AddCommand(newCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Open config in your $EDITOR",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.EditHandler,
	}
	editCmd.Flags().BoolP("verbose", "v", false, "")
	rootCmd.AddCommand(editCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all domain names saved in config",
		Args:  cobra.ExactArgs(0),
		Run:   handlers.ListHandler,
	}
	listCmd.Flags().BoolP("verbose", "v", false, "Also list all information about each domain")
	rootCmd.AddCommand(listCmd)

	switchCmd := &cobra.Command{
		Use:               "switch [name]",
		Short:             "Sets the active domain to the specified name",
		Example:           "apix switch myapi",
		Args:              cobra.ExactArgs(1),
		Run:               handlers.SwitchHandler,
		ValidArgsFunction: getDomainNames,
	}
	rootCmd.AddCommand(switchCmd)

	removeCmd := &cobra.Command{
		Use:               "remove [name]",
		Short:             "Remove a domain from the config",
		Example:           "apix remove myapi",
		Args:              cobra.ExactArgs(1),
		Run:               handlers.RemoveHandler,
		ValidArgsFunction: getDomainNames,
	}
	rootCmd.AddCommand(removeCmd)

	rootCmd.AddCommand(newDiscoverCommand())

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	registerHTTPCommands(rootCmd)
	registerGRPCCommands(rootCmd)

	return rootCmd
}

func getDomainNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	domains := config.GetDomainNames()
	return domains, cobra.ShellCompDirectiveNoFileComp
}

func expandConfigPath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to resolve home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
}
