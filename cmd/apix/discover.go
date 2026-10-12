package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rhydianjenkins/apix/v2/pkg/config"
	"github.com/rhydianjenkins/apix/v2/pkg/grpcclient"
	"github.com/rhydianjenkins/apix/v2/pkg/oas"
	"github.com/spf13/cobra"
)

func newDiscoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "List the endpoints or gRPC methods available on the active domain",
		Long: `Lists what can be called on the active domain without needing shell tab
completion - useful for scripts and non-interactive callers.

  - HTTP domains print "METHOD /path - summary" for each operation in the
    connected OpenAPI spec (see also: apix show <path> for full detail).
  - gRPC domains print "package.Service/Method(RequestType) -> ResponseType"
    for each method found via server reflection.`,
		Example:      "apix discover\napix discover --method post",
		Args:         cobra.ExactArgs(0),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			activeDomain := config.GetActiveDomain()

			if activeDomain.IsGRPC() {
				return discoverGRPC(activeDomain)
			}

			method, _ := cmd.Flags().GetString("method")
			return discoverHTTP(activeDomain, method)
		},
	}

	cmd.Flags().StringP("method", "m", "", "filter HTTP endpoints by verb (get, post, put, patch, delete); ignored for grpc domains")
	return cmd
}

func discoverHTTP(domain *config.Domain, method string) error {
	if !oas.HasValidOpenAPISpec(domain) {
		return fmt.Errorf("no OpenAPI spec is connected to domain %q; connect one with `apix new --oas <path/url>` or `apix edit`", domain.Name)
	}

	endpoints, err := oas.ListEndpoints(method, domain.HTTP.OpenAPISpecPath)

	if err != nil {
		return fmt.Errorf("failed to list endpoints: %w", err)
	}

	if len(endpoints) == 0 {
		fmt.Println("No endpoints found.")
		return nil
	}

	for _, e := range endpoints {
		line := fmt.Sprintf("%-7s %s", e.Method, e.Path)
		if e.Summary != "" {
			line += " - " + e.Summary
		}
		fmt.Println(line)
	}

	return nil
}

func discoverGRPC(domain *config.Domain) error {
	conn, err := grpcclient.Dial(domain)

	if err != nil {
		return fmt.Errorf("failed to dial %q: %w", domain.Name, err)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	methods, err := grpcclient.ListMethodsDetailed(ctx, conn, domain.Headers)

	if err != nil {
		return fmt.Errorf("failed to list methods via reflection: %w", err)
	}

	if len(methods) == 0 {
		fmt.Println("No methods found.")
		return nil
	}

	for _, m := range methods {
		fmt.Printf("%s(%s) -> %s\n", m.QualifiedName(), m.Input.FullName(), m.Output.FullName())
	}

	return nil
}
