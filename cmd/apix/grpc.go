package main

import (
	"fmt"
	"os"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/handlers"
	"github.com/spf13/cobra"
)

func addGRPCNewFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("insecure", false, "use plaintext instead of TLS for grpc connections (grpc only)")
}

func registerGRPCCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newGRPCCommand())
}

func newGRPCCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grpc [method]",
		Short:   "Invoke a unary gRPC method (with an empty request) on the active domain",
		Example: "apix grpc GetUser\napix grpc mypkg.UserService/GetUser",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			activeDomain := config.GetActiveDomain()

			if !activeDomain.IsGRPC() {
				fmt.Fprintf(os.Stderr, "active domain %q is not configured for grpc (protocol %q). Use `apix new --protocol grpc` or `apix edit`.\n", activeDomain.Name, activeDomain.Protocol)
				os.Exit(1)
			}

			headers, _ := cmd.Flags().GetStringSlice("header")
			headerMap := handlers.ParseHeaders(headers)

			body, err := handlers.GRPCHandler(activeDomain, args[0], headerMap)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making grpc call: %v\n", err)
				return
			}

			fmt.Printf("%s", string(body))
		},
	}

	cmd.Flags().StringSliceP("header", "H", []string{}, "Custom metadata in format 'Key: Value' (can be used multiple times)")
	return cmd
}
