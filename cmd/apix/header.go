package main

import (
	"github.com/rhydianjenkins/apix/v2/pkg/handlers"
	"github.com/spf13/cobra"
)

func newSetHeaderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-header <key> <value>",
		Short:        "Set a header/metadata value on a domain, without changing the active domain",
		Example:      "apix set-header authorization \"Bearer xyz\"\napix set-header authorization \"Bearer xyz\" --domain my-domain",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE:         handlers.SetHeaderHandler,
	}

	cmd.Flags().String("domain", "", "domain to set the header on (defaults to $APIX_DOMAIN, then the active domain)")

	return cmd
}
