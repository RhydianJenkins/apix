// This file owns every http-protocol command: the verb commands (get, post,
// put, patch, delete) and `show` (OpenAPI-spec browsing, which only makes
// sense for http domains). Nothing outside this file should know how http
// commands are built - root.go only calls registerHTTPCommands and
// addHTTPNewFlags.
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/handlers"
	"github.com/rhydianjenkins/apix/pkg/oas"
	"github.com/spf13/cobra"
)

// addHTTPNewFlags attaches the flags relevant only to http-protocol domains
// onto the shared `apix new` command.
func addHTTPNewFlags(cmd *cobra.Command) {
	cmd.Flags().String("user", "", "basic auth username to use for this domain (http only)")
	cmd.Flags().String("pass", "", "basic auth password to use for this domain (http only)")
	cmd.Flags().String("oas", "", "path to the oas spec for this endpoint (http only)")
}

// registerHTTPCommands adds every http-protocol command to rootCmd.
func registerHTTPCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newShowCommand())

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		rootCmd.AddCommand(createHTTPCommand(method))
	}
}

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "show [path]",
		Short:   "Show a summary of an endpoint from the connected OpenAPI spec",
		Example: "apix show /users/{id}",
		Args:    cobra.ExactArgs(1),
		Run:     handlers.ShowHandler,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain := config.GetActiveDomain()

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
}

func createHTTPCommand(method string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [path]", strings.ToLower(method)),
		Short:   fmt.Sprintf("Send a %s request to the active domain", method),
		Example: fmt.Sprintf("apix %s /users/123\ncat req_body.json | apix %s /users/123", strings.ToLower(method), strings.ToLower(method)),
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			activeDomain := config.GetActiveDomain()

			if activeDomain.IsGRPC() {
				fmt.Fprintf(os.Stderr, "active domain %q is configured for grpc, not http. Use `apix grpc` instead.\n", activeDomain.Name)
				os.Exit(1)
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

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making %s request: %v\n", method, err)
			}

			fmt.Printf("%s", string(body))
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain := config.GetActiveDomain()

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
