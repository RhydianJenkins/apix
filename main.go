package main
import (
	"fmt"
	"os"

	"github.com/rhydianjenkins/apix/pkg/handlers"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = initCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Short: "API eXecuter (APIX) is a CLI tool to manage API domains and make requests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var setCmd = &cobra.Command{
		Use: "set [name] [base]",
		Short: "Set a new or existing API domain",
		Args: cobra.ExactArgs(2),
		Run: handlers.SetHandler,
	}
	setCmd.Flags().String("user", "", "basic auth username to use for this domain")
	setCmd.Flags().String("pass", "", "basic auth password to use for this domain")
	rootCmd.AddCommand(setCmd)

	var listCmd = &cobra.Command{
		Use: "list",
		Short: "List all domain names saved in config",
		Args: cobra.ExactArgs(0),
		Run: handlers.ListHandler,
	}
	listCmd.Flags().Bool("verbose", false, "Also list all information about each domain")
	rootCmd.AddCommand(listCmd)

	// TODO Rhydian more commands...
	// post, patch, put, etc... may have to think more on the structure of this one?
	var getCmd = &cobra.Command{
		Use: "get [path]",
		Short: "Send a GET request to the active domain",
		Args: cobra.MaximumNArgs(2),
		Run: handlers.GetHandler,
	}
	rootCmd.AddCommand(getCmd)

	var useCmd = &cobra.Command{
		Use: "use [name]",
		Short: "Sets the active domain to the specified name",
		Args: cobra.MinimumNArgs(1),
		Run: handlers.UseHandler,
	}
	rootCmd.AddCommand(useCmd)

	var removeCmd = &cobra.Command{
		Use: "remove [name]",
		Short: "Remove a domain from the config",
		Args: cobra.MinimumNArgs(1),
		Run: handlers.RemoveHandler,
	}
	rootCmd.AddCommand(removeCmd)

	return rootCmd
}
