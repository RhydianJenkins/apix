package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/grpcclient"
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/rhydianjenkins/apix/v3/internal/oas"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [path|method]",
		Short: "List the endpoints or gRPC methods on the active domain, or show a summary of one",
		Long: `Without an argument, lists what can be called on the active domain without
needing shell tab completion - useful for scripts and non-interactive callers:

  - HTTP domains print "METHOD /path - summary" for each operation in the
    connected OpenAPI spec.
  - gRPC domains print "package.Service/Method(RequestType) -> ResponseType"
    for each method found via server reflection.

With a path or method name, shows a full summary of that single endpoint or
method instead.`,
		Example:      "apix show\napix show --method post\napix show /users/{id}\napix show GetUser\napix show mypkg.UserService/GetUser",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			activeDomain, err := config.RequireActiveDomain()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				if activeDomain.IsGRPC() {
					return listGRPCMethods(activeDomain)
				}

				method, _ := cmd.Flags().GetString("method")
				return listHTTPEndpoints(activeDomain, method)
			}

			handlers.ShowHandler(cmd, args)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain, err := config.RequireActiveDomain()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if activeDomain.IsGRPC() {
				conn, err := grpcclient.Dial(activeDomain)
				if err != nil {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}
				defer conn.Close()

				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				methods, err := grpcclient.ListMethods(ctx, conn, activeDomain.Headers)
				if err != nil {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}

				return methods, cobra.ShellCompDirectiveNoFileComp
			}

			if !oas.HasValidOpenAPISpec(activeDomain) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			endpoints, err := oas.GetEndpointsValidArgs("", activeDomain.HTTP.OpenAPISpecPath)

			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return endpoints, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceP("header", "H", []string{}, "Custom metadata in format 'Key: Value' (can be used multiple times, grpc only)")
	cmd.Flags().StringP("method", "m", "", "filter HTTP endpoints by verb (get, post, put, patch, delete); only applies when no path/method is given, ignored for grpc domains")
	return cmd
}

func listHTTPEndpoints(domain *config.Domain, method string) error {
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

func listGRPCMethods(domain *config.Domain) error {
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
