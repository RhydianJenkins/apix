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

	existingDomain, _ := config.LoadDomain(name)

	if existingDomain != nil {
		if domain.Base == "" {
			domain.Base = existingDomain.Base
		}
		if domain.Name == "" {
			domain.Name = existingDomain.Name
		}
		if domain.User == "" {
			domain.User = existingDomain.User
		}
		if domain.Pass == "" {
			domain.Pass = existingDomain.Pass
		}
	}

	config.SetDomain(domain)
}
