package main

import (
	"github.com/rhydianjenkins/apix/pkg/handlers"
	"github.com/spf13/cobra"
)

func addAuthNewFlags(cmd *cobra.Command) {
	cmd.Flags().String("auth-command", "", "shell command `apix auth` should run to (re-)authenticate this domain")
	cmd.Flags().String("auth-header", "", "if set, the header/metadata key to store the auth command's trimmed stdout under")
}

func registerAuthCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newAuthCommand())
}

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "auth [name]",
		Short:             "Run a domain's configured auth command (see `auth.command` in config)",
		Example:           "apix auth\napix auth junifer-uat\napix auth --all",
		Args:              cobra.MaximumNArgs(1),
		SilenceUsage:      true,
		RunE:              handlers.AuthHandler,
		ValidArgsFunction: getDomainNames,
	}

	cmd.Flags().Bool("all", false, "run every domain's configured auth command")

	return cmd
}
