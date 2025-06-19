package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func GetHandler(cmd *cobra.Command, args []string) {
	path := "/"

	if len(args) > 0 {
		path = args[0]
	}

	domain, err := config.LoadActiveDomain()

	if err != nil {
		println("Error loading domain:", err)
	}

	fmt.Printf("TODO Rhydian GET curl request\n")
	fmt.Printf("Domain: %s\n", domain.Name)
	fmt.Printf("Path: %s\n", path)
}
