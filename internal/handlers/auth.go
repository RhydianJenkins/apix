package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/cobra"
)

func AuthHandler(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")

	if all {
		if len(args) > 0 {
			return fmt.Errorf("cannot use --all together with a domain name")
		}
		return authAll()
	}

	var domain *config.Domain
	var err error

	if len(args) == 1 {
		domain, err = config.LoadDomain(args[0])
		if err != nil {
			return err
		}
	} else {
		domain = config.GetActiveDomain()
	}

	return runAuth(domain)
}

func authAll() error {
	names := config.GetDomainNames()
	sort.Strings(names)

	var failed []string

	for _, name := range names {
		domain, err := config.LoadDomain(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			failed = append(failed, name)
			continue
		}

		if !domain.HasAuth() {
			continue
		}

		if err := runAuth(domain); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			failed = append(failed, name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("auth failed for: %s", strings.Join(failed, ", "))
	}

	return nil
}

func runAuth(domain *config.Domain) error {
	if !domain.HasAuth() {
		return fmt.Errorf(
			"domain %q has no auth command configured. Set one with `apix edit` (add an `auth.command` key) or `apix new ... --auth-command '<cmd>'`",
			domain.Name,
		)
	}

	execCmd := exec.Command("sh", "-c", domain.Auth.Command)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// set env variables so child apix processes know which domain/config it was ran from
	// this allows you to call apix from within `auth.command`
	execCmd.Env = append(os.Environ(),
		"APIX_DOMAIN="+domain.Name,
		"APIX_CONFIG="+strings.Join(config.CfgPaths, string(os.PathListSeparator)),
	)

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("auth command for %q failed: %w", domain.Name, err)
	}

	fmt.Printf("Ran auth command for %q.\n", domain.Name)

	return nil
}
