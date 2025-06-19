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
		Short: "apix is a CLI tool to manage API domains and make requests",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var setCmd = &cobra.Command{
		Use:   "set [name] [base]",
		Short: "Set a new or existing API domain",
		Args:  cobra.ExactArgs(2),
		Run: handlers.SetHandler,
	}
	setCmd.Flags().String("user", "", "basic auth username to use for this domain")
	setCmd.Flags().String("pass", "", "basic auth password to use for this domain")
	rootCmd.AddCommand(setCmd)

	return rootCmd
}
