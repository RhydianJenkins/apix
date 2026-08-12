package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/oas"
	"github.com/spf13/cobra"
)

func ShowHandler(cmd *cobra.Command, args []string) {
	path := args[0]
	domain := config.GetActiveDomain()

	if !oas.HasValidOpenAPISpec(domain) {
		fmt.Printf("No OpenAPI spec is connected to the active domain %q.\nConnect one with `apix new --oas <path/url>` or `apix edit`.\n", domain.Name)
		return
	}

	pathItem, err := oas.GetPathItem(path, domain.OpenAPISpecPath)

	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Print(oas.FormatPathItemSummary(path, pathItem))
}
