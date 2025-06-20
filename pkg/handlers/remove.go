package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func RemoveHandler(cmd *cobra.Command, args []string) {
	domainToRemove := args[0]

	if err := config.RemoveDomain(domainToRemove); err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Printf("Domain %q removed successfully.\n", domainToRemove)
}
