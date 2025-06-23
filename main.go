package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/handlers"
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
		Use: "apix",
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
	rootCmd.AddCommand(newCmd)

	var listCmd = &cobra.Command{
		Use: "list",
		Short: "List all domain names saved in config",
		Args: cobra.ExactArgs(0),
		Run: handlers.ListHandler,
	}
	listCmd.Flags().Bool("verbose", false, "Also list all information about each domain")
	rootCmd.AddCommand(listCmd)

	var useCmd = &cobra.Command{
		Use: "use [name]",
		Short: "Sets the active domain to the specified name",
		Example: "apix use myapi",
		Args: cobra.MinimumNArgs(1),
		Run: handlers.UseHandler,
		// TODO RHydian complete with a list of available domains in config
		// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// 	suggestions := []string{"test1", "test2", "test3"}
		// 	return suggestions, cobra.ShellCompDirectiveNoFileComp
		// },
	}
	rootCmd.AddCommand(useCmd)

	var removeCmd = &cobra.Command{
		Use: "remove [name]",
		Short: "Remove a domain from the config",
		Example: "apix remove myapi",
		Args: cobra.MinimumNArgs(1),
		Run: handlers.RemoveHandler,
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
    return &cobra.Command{
		Use: fmt.Sprintf("%s [path] [body]", strings.ToLower(method)),
        Short: fmt.Sprintf("Send a %s request to the active domain", method),
        Example: fmt.Sprintf("apix %s /users/123", strings.ToLower(method)),
        Args: cobra.RangeArgs(1, 2),
        Run: func(cmd *cobra.Command, args []string) {
			input, _ := getStdIn()

			body, err := handlers.HTTPHandler(
				method,
				config.LoadActiveDomain(),
				args[0],
				input,
				nil, // TODO Rhydian how to specify this?
			)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making %s request: %v\n", method, err)
			}

			fmt.Printf("%s", string(body))
        },
    }
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
