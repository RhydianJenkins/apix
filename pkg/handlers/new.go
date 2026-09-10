package handlers

import (
	"fmt"
	"os"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func NewHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	base := args[1]
	user, _ := cmd.Flags().GetString("user")
	pass, _ := cmd.Flags().GetString("pass")
	oas, _ := cmd.Flags().GetString("oas")
	protocol, _ := cmd.Flags().GetString("protocol")
	insecure, _ := cmd.Flags().GetBool("insecure")
	headers, _ := cmd.Flags().GetStringSlice("header")
	headerMap := ParseHeaders(headers)

	if protocol != config.ProtocolHTTP && protocol != config.ProtocolGRPC {
		fmt.Fprintf(os.Stderr, "invalid --protocol %q: must be %q or %q\n", protocol, config.ProtocolHTTP, config.ProtocolGRPC)
		os.Exit(1)
	}

	var domain = &config.Domain{
		Base:     base,
		Name:     name,
		Protocol: protocol,
		Headers:  headerMap,
	}

	if protocol == config.ProtocolGRPC {
		domain.GRPC = &config.GRPCOptions{
			Insecure: insecure,
		}
	} else {
		domain.HTTP = &config.HTTPOptions{
			User:            user,
			Pass:            pass,
			OpenAPISpecPath: oas,
		}
	}

	existingDomain, _ := config.LoadDomain(name)

	if existingDomain != nil {
		if domain.Base == "" {
			domain.Base = existingDomain.Base
		}
		if domain.Name == "" {
			domain.Name = existingDomain.Name
		}
		if domain.Headers == nil {
			domain.Headers = existingDomain.Headers
		}

		if domain.HTTP != nil && existingDomain.HTTP != nil {
			if domain.HTTP.User == "" {
				domain.HTTP.User = existingDomain.HTTP.User
			}
			if domain.HTTP.Pass == "" {
				domain.HTTP.Pass = existingDomain.HTTP.Pass
			}
			if domain.HTTP.OpenAPISpecPath == "" {
				domain.HTTP.OpenAPISpecPath = existingDomain.HTTP.OpenAPISpecPath
			}
		}
	}

	config.SetDomain(domain)

	println("New domain added and active.")
}
