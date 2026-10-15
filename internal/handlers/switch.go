package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/cobra"
)

func SwitchHandler(cmd *cobra.Command, args []string) {
	newActiveDomain := args[0]

	err := config.SetActiveName(newActiveDomain)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Printf("Switched to domain '%s'.\n", newActiveDomain)
}
