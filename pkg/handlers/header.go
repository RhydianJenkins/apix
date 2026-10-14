package handlers

import (
	"fmt"
	"os"

	"github.com/rhydianjenkins/apix/v2/pkg/config"
	"github.com/spf13/cobra"
)

// resolveDomainName picks the domain a command should act on: an explicit
// --domain flag wins, then the APIX_DOMAIN environment variable (set by
// `apix auth` when it runs a domain's auth command), then the active domain.
func resolveDomainName(cmd *cobra.Command) string {
	if domain, _ := cmd.Flags().GetString("domain"); domain != "" {
		return domain
	}

	if domain := os.Getenv("APIX_DOMAIN"); domain != "" {
		return domain
	}

	return config.GetActiveDomain().Name
}

// SetHeaderHandler persists a single header/metadata value on a domain
// without changing the active domain. It's the mechanism an `auth.command`
// uses to save whatever credential it mints - apix itself never inspects or
// captures the command's output.
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
