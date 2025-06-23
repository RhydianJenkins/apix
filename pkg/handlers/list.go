package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func ListHandler(cmd *cobra.Command, args []string) {
	var config = config.LoadConfig()
	var verbose, _ = cmd.Flags().GetBool("verbose")

	for name := range config.Domains {
		marker := ""
		if name == config.Active {
			marker = " *"
		}

		fmt.Println(name + marker)

		if verbose {
			domain := config.Domains[name]
			fmt.Printf("\tBase: %s\n", domain.Base)
			fmt.Printf("\tUser: %s\n", domain.User)
			fmt.Printf("\tPass: %s\n", domain.Pass)
			fmt.Printf("\tOAS: %s\n", domain.OpenAPISpecPath)
		}
	}
}
