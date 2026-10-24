package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/cobra"
)

func StatusHandler(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if cfg.Active == "" {
		return fmt.Errorf("no active domain set. Use `apix switch` to set one, or `apix new` to create one")
	}

	domain, ok := cfg.Domains[cfg.Active]
	if !ok {
		return fmt.Errorf("active domain %q is no longer defined in config. Use `apix switch` to set a valid one", cfg.Active)
	}
	domain.Name = cfg.Active

	fmt.Println("Active domain: " + domain.Name)

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		printDomainDetails(&domain)
	}

	return nil
}
