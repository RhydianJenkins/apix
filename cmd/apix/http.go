package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
	"github.com/rhydianjenkins/apix/v3/internal/handlers"
	"github.com/rhydianjenkins/apix/v3/internal/oas"
	"github.com/spf13/cobra"
)

func addHTTPNewFlags(cmd *cobra.Command) {
	cmd.Flags().String("user", "", "basic auth username to use for this domain (http only)")
	cmd.Flags().String("pass", "", "basic auth password to use for this domain (http only)")
	cmd.Flags().String("oas", "", "path to the oas spec for this endpoint (http only)")
	cmd.Flags().Bool("graphql", false, "treat this domain as GraphQL: discover operations for `apix show` via introspection instead of an OAS spec (http only)")
}

func registerHTTPCommands(rootCmd *cobra.Command) {
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		rootCmd.AddCommand(createHTTPCommand(method))
	}
}

func createHTTPCommand(method string) *cobra.Command {
	// GraphQL invocation only applies to POST - the only verb GraphQL servers use.
	isPost := method == "POST"

	cmd := &cobra.Command{
		Use:          fmt.Sprintf("%s [path]", strings.ToLower(method)),
		Short:        fmt.Sprintf("Send a %s request to the active domain", method),
		Example:      httpCommandExample(method, isPost),
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

			path := args[0]

			var body []byte

			if isPost && activeDomain.IsGraphQL() && path != "" && path != "/" {
				fields, _ := cmd.Flags().GetStringSlice("fields")
				body, err = handlers.GraphQLPostHandler(activeDomain, path, fields, input, headerMap)
			} else {
				body, err = handlers.HTTPHandler(method, activeDomain, path, input, headerMap)
			}

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

			if isPost && activeDomain.IsGraphQL() {
				infos, err := graphqlclient.ListOperations(activeDomain, activeDomain.Headers)
				if err != nil {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}

				names := make([]string, len(infos))
				for i, info := range infos {
					names[i] = info.QualifiedName()
				}

				return names, cobra.ShellCompDirectiveNoFileComp
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

	if isPost {
		cmd.Flags().StringSlice("fields", nil, "extend a GraphQL operation's default field selection with dotted paths into nested fields (e.g. account.energy.properties), relative to its return type; repeatable (GraphQL domains only)")
	}

	return cmd
}

func httpCommandExample(method string, isPost bool) string {
	lower := strings.ToLower(method)
	example := fmt.Sprintf("apix %s /users/123\ncat req_body.json | apix %s /users/123", lower, lower)

	if isPost {
		example += "\n\n# on a --graphql domain, post can invoke a discovered operation directly,\n# with stdin as its variables, instead of hand-building a query document:\necho '{\"id\": \"123\"}' | apix post Query.user"
	}

	return example
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
