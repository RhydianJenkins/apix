package oas

import (
	"fmt"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// FormatPathItemSummary renders a human-readable summary of every operation
// defined for a path, suitable for printing directly to the terminal.
func FormatPathItemSummary(path string, pathItem *v3.PathItem) string {
	var b strings.Builder

	if pathItem.Summary != "" {
		fmt.Fprintf(&b, "%s\n", pathItem.Summary)
	}

	if pathItem.Description != "" {
		fmt.Fprintf(&b, "%s\n", pathItem.Description)
	}

	if pathItem.Summary != "" || pathItem.Description != "" {
		fmt.Fprintln(&b)
	}

	operations := pathItem.GetOperations()
	first := true

	for method, operation := range operations.FromOldest() {
		if !first {
			fmt.Fprintln(&b)
		}
		first = false

		formatOperation(&b, strings.ToUpper(method), path, operation)
	}

	return b.String()
}

func formatOperation(b *strings.Builder, method, path string, operation *v3.Operation) {
	fmt.Fprintf(b, "%s %s\n", method, path)

	if operation.Summary != "" {
		fmt.Fprintf(b, "\tSummary: %s\n", operation.Summary)
	}

	if operation.Description != "" {
		fmt.Fprintf(b, "\tDescription: %s\n", operation.Description)
	}

	if operation.OperationId != "" {
		fmt.Fprintf(b, "\tOperation ID: %s\n", operation.OperationId)
	}

	if len(operation.Tags) > 0 {
		fmt.Fprintf(b, "\tTags: %s\n", strings.Join(operation.Tags, ", "))
	}

	if operation.Deprecated != nil && *operation.Deprecated {
		fmt.Fprintf(b, "\tDeprecated: true\n")
	}

	if len(operation.Parameters) > 0 {
		fmt.Fprintf(b, "\tParameters:\n")
		for _, param := range operation.Parameters {
			fmt.Fprintf(b, "\t\t%s\n", formatParameter(param))
		}
	}

	if operation.RequestBody != nil {
		fmt.Fprintf(b, "\tRequest Body:\n")
		fmt.Fprintf(b, "\t\tRequired: %t\n", operation.RequestBody.Required != nil && *operation.RequestBody.Required)

		if operation.RequestBody.Description != "" {
			fmt.Fprintf(b, "\t\tDescription: %s\n", operation.RequestBody.Description)
		}

		if operation.RequestBody.Content != nil {
			contentTypes := []string{}
			for contentType := range operation.RequestBody.Content.KeysFromOldest() {
				contentTypes = append(contentTypes, contentType)
			}
			fmt.Fprintf(b, "\t\tContent-Types: %s\n", strings.Join(contentTypes, ", "))
		}
	}

	if operation.Responses != nil {
		fmt.Fprintf(b, "\tResponses:\n")

		for code, response := range operation.Responses.Codes.FromOldest() {
			fmt.Fprintf(b, "\t\t%s: %s\n", code, response.Description)
		}

		if operation.Responses.Default != nil {
			fmt.Fprintf(b, "\t\tdefault: %s\n", operation.Responses.Default.Description)
		}
	}
}

func formatParameter(param *v3.Parameter) string {
	required := false
	if param.Required != nil {
		required = *param.Required
	}

	details := param.In
	if required {
		details += ", required"
	} else {
		details += ", optional"
	}

	paramType := ""
	if param.Schema != nil {
		if schema := param.Schema.Schema(); schema != nil && len(schema.Type) > 0 {
			paramType = strings.Join(schema.Type, ", ")
		}
	}

	summary := fmt.Sprintf("%s (%s)", param.Name, details)

	if paramType != "" {
		summary += fmt.Sprintf(" %s", paramType)
	}

	if param.Description != "" {
		summary += fmt.Sprintf(" - %s", param.Description)
	}

	return summary
}
