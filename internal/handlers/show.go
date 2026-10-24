package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/oas"
	"github.com/spf13/cobra"
)

func ShowHandler(cmd *cobra.Command, args []string) {
	target := args[0]
	domain, err := config.RequireActiveDomain()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	if domain.IsGRPC() {
		headers, _ := cmd.Flags().GetStringSlice("header")
		headerMap := ParseHeaders(headers)

		summary, err := GRPCShowHandler(domain, target, headerMap)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		fmt.Print(summary)
		return
	}

	if !oas.HasValidOpenAPISpec(domain) {
		fmt.Printf("No OpenAPI spec is connected to the active domain %q.\nConnect one with `apix new --oas <path/url>` or `apix edit`.\n", domain.Name)
		return
	}

	pathItem, err := oas.GetPathItem(target, domain.HTTP.OpenAPISpecPath)

	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Print(oas.FormatPathItemSummary(target, pathItem))
}
