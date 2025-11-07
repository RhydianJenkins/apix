package main

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

func main() {
	var rootCmd = initCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use: "apix [command]",
		Short: "API eXecuter (APIX) is a CLI tool to manage API domains and make requests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var newCmd = &cobra.Command{
		Use: "new [name] [base]",
		Short: "Create a new API domain",
		Example: "apix new myapi https://api.example.com --user foo --pass bar",
		Args: cobra.ExactArgs(2),
		Run: handlers.NewHandler,
	}
	newCmd.Flags().String("user", "", "basic auth username to use for this domain")
	newCmd.Flags().String("pass", "", "basic auth password to use for this domain")
	newCmd.Flags().String("oas", "", "path to the oas spec for this endpoint")
	newCmd.Flags().StringSliceP("header", "H", []string{}, "default headers for this domain in format 'Key: Value' (can be used multiple times)")
	rootCmd.AddCommand(newCmd)

	var editCmd = &cobra.Command{
		Use: "edit",
		Short: "Open config in your $EDITOR",
		Args: cobra.ExactArgs(0),
		Run: handlers.EditHandler,
	}
	editCmd.Flags().Bool("verbose", false, "")
	rootCmd.AddCommand(editCmd)

	var listCmd = &cobra.Command{
		Use: "list",
		Short: "List all domain names saved in config",
		Args: cobra.ExactArgs(0),
		Run: handlers.ListHandler,
	}
	listCmd.Flags().Bool("verbose", false, "Also list all information about each domain")
	rootCmd.AddCommand(listCmd)

	var switchCmd = &cobra.Command{
		Use: "switch [name]",
		Short: "Sets the active domain to the specified name",
		Example: "apix switch myapi",
		Args: cobra.OnlyValidArgs,
		Run: handlers.SwitchHandler,
		ValidArgsFunction: getDomainNames,
	}
	rootCmd.AddCommand(switchCmd)

	var removeCmd = &cobra.Command{
		Use: "remove [name]",
		Short: "Remove a domain from the config",
		Example: "apix remove myapi",
		Args: cobra.OnlyValidArgs,
		Run: handlers.RemoveHandler,
		ValidArgsFunction: getDomainNames,
	}
	rootCmd.AddCommand(removeCmd)

	rootCmd.AddCommand(createHTTPCommand("GET"))
	rootCmd.AddCommand(createHTTPCommand("POST"))
	rootCmd.AddCommand(createHTTPCommand("PUT"))
	rootCmd.AddCommand(createHTTPCommand("PATCH"))
	rootCmd.AddCommand(createHTTPCommand("DELETE"))

	return rootCmd
}

func createHTTPCommand(method string) *cobra.Command {
    cmd := &cobra.Command{
		Use: fmt.Sprintf("%s [path]", strings.ToLower(method)),
        Short: fmt.Sprintf("Send a %s request to the active domain", method),
        Example: fmt.Sprintf("apix %s /users/123\ncat req_body.json | apix %s /users/123", strings.ToLower(method), strings.ToLower(method)),
        Args: cobra.RangeArgs(1, 2),
        Run: func(cmd *cobra.Command, args []string) {
			input, _ := getStdIn()
			headers, _ := cmd.Flags().GetStringSlice("header")
			headerMap := handlers.ParseHeaders(headers)

			body, err := handlers.HTTPHandler(
				method,
				config.GetActiveDomain(),
				args[0],
				input,
				headerMap,
			)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making %s request: %v\n", method, err)
			}

			fmt.Printf("%s", string(body))
        },
		ValidArgsFunction: func (cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			activeDomain := config.GetActiveDomain()

			if !oas.HasValidOpenAPISpec(activeDomain) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			oasPath := activeDomain.OpenAPISpecPath
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

func getDomainNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	domains := config.GetDomainNames()
	return domains, cobra.ShellCompDirectiveNoFileComp
}
