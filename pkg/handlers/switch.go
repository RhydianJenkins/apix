package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func SwitchHandler(cmd *cobra.Command, args []string) {
	newActiveDomain := args[0]

	err := config.SetActiveDomain(newActiveDomain)
	if err != nil {
		fmt.Printf("Error setting active domain: %v\n", err)
		return
	}

	fmt.Printf("Switched active domain to '%s'.\n", newActiveDomain)
}
