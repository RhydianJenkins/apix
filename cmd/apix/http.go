package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/rhydianjenkins/apix/v3/internal/oas"
	"github.com/spf13/cobra"
)

func addHTTPNewFlags(cmd *cobra.Command) {
	cmd.Flags().String("user", "", "basic auth username to use for this domain (http only)")
	cmd.Flags().String("pass", "", "basic auth password to use for this domain (http only)")
	cmd.Flags().String("oas", "", "path to the oas spec for this endpoint (http only)")
}

func registerHTTPCommands(rootCmd *cobra.Command) {
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		rootCmd.AddCommand(createHTTPCommand(method))
	}
}

func createHTTPCommand(method string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          fmt.Sprintf("%s [path]", strings.ToLower(method)),
		Short:        fmt.Sprintf("Send a %s request to the active domain", method),
		Example:      fmt.Sprintf("apix %s /users/123\ncat req_body.json | apix %s /users/123", strings.ToLower(method), strings.ToLower(method)),
		Args:         cobra.RangeArgs(1, 2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			activeDomain, err := config.RequireActiveDomain()
			if err != nil {
				return err
			}

			if activeDomain.IsGRPC() {
				return fmt.Errorf("active domain %q is configured for grpc, not http. Use `apix grpc` instead", activeDomain.Name)
			}

			input, _ := getStdIn()
			headers, _ := cmd.Flags().GetStringSlice("header")
			headerMap := handlers.ParseHeaders(headers)

			body, err := handlers.HTTPHandler(
				method,
				activeDomain,
				args[0],
				input,
				headerMap,
			)

			fmt.Printf("%s", string(body))

			if err != nil {
				return fmt.Errorf("error making %s request: %w", method, err)
			}

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain, err := config.RequireActiveDomain()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if !oas.HasValidOpenAPISpec(activeDomain) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			oasPath := activeDomain.HTTP.OpenAPISpecPath
			endpoints, err := oas.GetEndpointsValidArgs(method, oasPath)

			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return endpoints, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceP("header", "H", []string{}, "Custom headers in format 'Key: Value' (can be used multiple times)")
	return cmd
}

func getStdIn() (*[]byte, error) {
	stat, err := os.Stdin.Stat()

	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		return &input, err
	}

	return nil, nil
}
