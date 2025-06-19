package handlers

import (
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func SetHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	base := args[1]
	user, _ := cmd.Flags().GetString("user")
	pass, _ := cmd.Flags().GetString("pass")

	var domain = &config.Domain{
		Base: base,
		Name: name,
		User: user,
		Pass: pass,
	}

	config.SetDomain(domain)
}
