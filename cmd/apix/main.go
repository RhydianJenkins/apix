package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// version is overridden at build time via -ldflags "-X main.version=...".
// See the `build`/`release` apps and the `ldflags` in flake.nix.
var version = "dev"

func main() {
	if err := newRootCmd(resolveVersion(version)).Execute(); err != nil {
		os.Exit(1)
	}
}

func resolveVersion(v string) string {
	if v != "dev" {
		return v
	}

	// ld flags not set with built with `go install ...`
	// we should resolve it with the VCS metadata in build info instead
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return v
	}

	if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}

	var revision string
	var modified bool
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}

	if revision == "" {
		return v
	}

	if len(revision) > 12 {
		revision = revision[:12]
	}

	if modified {
		return fmt.Sprintf("dev-%s-dirty", revision)
	}

	return fmt.Sprintf("dev-%s", revision)
}

func newRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "apix [command]",
		Short: "API eXecuter (APIX) is a CLI tool to manage API domains and make requests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.PersistentFlags().VarP(&cfgPathValue{}, "config", "c", "path to an apix config file; repeat to layer files, later ones override earlier ones and receive all writes")

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
	addAuthNewFlags(newCmd)
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

	rootCmd.AddCommand(newShowCommand())

	registerHTTPCommands(rootCmd)
	registerGRPCCommands(rootCmd)
	registerAuthCommands(rootCmd)

	return rootCmd
}

func getDomainNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	domains := config.GetDomainNames()
	return domains, cobra.ShellCompDirectiveNoFileComp
}

// cfgPathValue collects repeated -c flags into config.CfgPaths. The first
// occurrence replaces the default list, later ones append.
type cfgPathValue struct{ set bool }

func (c *cfgPathValue) String() string { return strings.Join(config.CfgPaths, ",") }
func (c *cfgPathValue) Type() string   { return "stringArray" }

func (c *cfgPathValue) Set(path string) error {
	expanded, err := config.ExpandPath(path)
	if err != nil {
		return err
	}

	if !c.set {
		config.CfgPaths = nil
		c.set = true
	}

	config.CfgPaths = append(config.CfgPaths, expanded)

	return nil
}

var _ pflag.Value = (*cfgPathValue)(nil)
