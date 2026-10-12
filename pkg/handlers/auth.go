package handlers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/rhydianjenkins/apix/v2/pkg/config"
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

// Stdin/stderr are inherited so interactive auth flows (device-code OAuth,
// "approve on your phone" prompts) work as expected.
func runAuth(domain *config.Domain) error {
	if !domain.HasAuth() {
		return fmt.Errorf(
			"domain %q has no auth command configured. Set one with `apix edit` (add an `auth.command` key) or `apix new ... --auth-command '<cmd>'`",
			domain.Name,
		)
	}

	execCmd := exec.Command("sh", "-c", domain.Auth.Command)
	execCmd.Stdin = os.Stdin
	execCmd.Stderr = os.Stderr

	header := domain.Auth.Header

	var out bytes.Buffer
	if header != "" {
		execCmd.Stdout = &out
	} else {
		execCmd.Stdout = os.Stdout
	}

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("auth command for %q failed: %w", domain.Name, err)
	}

	if header == "" {
		fmt.Printf("Ran auth command for %q.\n", domain.Name)
		return nil
	}

	value := strings.TrimSpace(out.String())
	if value == "" {
		return fmt.Errorf("auth command for %q configured header %q but produced no output", domain.Name, header)
	}

	if err := config.SetDomainHeader(domain.Name, header, value); err != nil {
		return fmt.Errorf("failed to save %q header for %q: %w", header, domain.Name, err)
	}

	fmt.Printf("Updated %q header for domain %q.\n", header, domain.Name)
	return nil
}
