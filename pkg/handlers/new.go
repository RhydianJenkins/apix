package handlers

import (
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func NewHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	base := args[1]
	user, _ := cmd.Flags().GetString("user")
	pass, _ := cmd.Flags().GetString("pass")
	oas, _ := cmd.Flags().GetString("oas")

	var domain = &config.Domain{
		Base: base,
		Name: name,
		User: user,
		Pass: pass,
		OpenAPISpecPath: oas,
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
		if domain.OpenAPISpecPath == "" {
			domain.OpenAPISpecPath = existingDomain.OpenAPISpecPath
		}
	}

	config.SetDomain(domain)

	println("New domain added and active.")
}
