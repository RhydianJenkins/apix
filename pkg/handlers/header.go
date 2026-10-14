package handlers

import (
	"fmt"
	"os"

	"github.com/rhydianjenkins/apix/v2/pkg/config"
	"github.com/spf13/cobra"
)

func resolveDomainName(cmd *cobra.Command) string {
	if domain, _ := cmd.Flags().GetString("domain"); domain != "" {
		return domain
	}

	// we set `APIX_DOMAIN` when running as a sub-process in `apix auth`
	if domain := os.Getenv("APIX_DOMAIN"); domain != "" {
		return domain
	}

	return config.GetActiveDomain().Name
}

func SetHeaderHandler(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	name := resolveDomainName(cmd)
	if name == "" {
		return fmt.Errorf("no domain to set the header on: pass --domain, run this from within `apix auth`, or set an active domain with `apix switch`")
	}

	if err := config.SetDomainHeader(name, key, value); err != nil {
		return fmt.Errorf("failed to set %q header for %q: %w", key, name, err)
	}

	fmt.Printf("Updated %q header for domain %q.\n", key, name)

	return nil
}
