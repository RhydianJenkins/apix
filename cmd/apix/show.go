package main

import (
	"context"
	"time"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/grpcclient"
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/rhydianjenkins/apix/v3/internal/oas"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [path|method]",
		Short:   "Show a summary of an HTTP endpoint (from the connected OpenAPI spec) or a gRPC method (via server reflection)",
		Example: "apix show /users/{id}\napix show GetUser\napix show mypkg.UserService/GetUser",
		Args:    cobra.ExactArgs(1),
		Run:     handlers.ShowHandler,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain := config.GetActiveDomain()

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
	return cmd
}
