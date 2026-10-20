package main

import (
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/spf13/cobra"
)

func addAuthNewFlags(cmd *cobra.Command) {
	cmd.Flags().String("auth-command", "", "shell command `apix auth` should run to (re-)authenticate this domain; responsible for persisting any header itself, e.g. via `apix set-header`")
}

func registerAuthCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newAuthCommand())
	rootCmd.AddCommand(newSetHeaderCommand())
}

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "auth [name]",
		Short:             "Run a domain's configured auth command (see `auth.command` in config)",
		Example:           "apix auth\napix auth my-api-domain\napix auth --all",
		Args:              cobra.MaximumNArgs(1),
		SilenceUsage:      true,
		RunE:              handlers.AuthHandler,
		ValidArgsFunction: getDomainNames,
	}

	cmd.Flags().BoolP("all", "a", false, "run every domain's configured auth command")

	return cmd
}
