package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func ListHandler(cmd *cobra.Command, args []string) {
	var cfg = config.LoadConfig()
	var verbose, _ = cmd.Flags().GetBool("verbose")

	for name := range cfg.Domains {
		marker := ""
		if name == cfg.Active {
			marker = " *"
		}

		fmt.Println(name + marker)

		if verbose {
			domain := cfg.Domains[name]
			protocol := domain.Protocol
			if protocol == "" {
				protocol = config.ProtocolHTTP
			}

			fmt.Printf("\tBase: %s\n", domain.Base)
			fmt.Printf("\tProtocol: %s\n", protocol)

			if domain.HTTP != nil {
				fmt.Printf("\tUser: %s\n", domain.HTTP.User)
				fmt.Printf("\tPass: %s\n", domain.HTTP.Pass)
				fmt.Printf("\tOAS: %s\n", domain.HTTP.OpenAPISpecPath)
				if domain.HTTP.Port != 0 {
					fmt.Printf("\tPort: %d\n", domain.HTTP.Port)
				}
			}

			if domain.GRPC != nil {
				fmt.Printf("\tInsecure: %t\n", domain.GRPC.Insecure)
				fmt.Printf("\tPort: %d\n", domain.GRPC.Port)
			}

			if len(domain.Headers) > 0 {
				fmt.Printf("\tHeaders:\n")
				for k, v := range domain.Headers {
					fmt.Printf("\t\t%s: %s\n", k, v)
				}
			}
		}
	}
}
