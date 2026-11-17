package graphqlclient

import (
	"fmt"
	"strings"
)

// FormatOperationSummary renders a human-readable summary of a single
// GraphQL operation (a Query/Mutation/Subscription field), without invoking
// it.
func FormatOperationSummary(info *OperationInfo) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n", info.Signature())

	if info.Description != "" {
		fmt.Fprintf(&b, "\tDescription: %s\n", info.Description)
	}

	if info.IsDeprecated {
		if info.DeprecationReason != "" {
			fmt.Fprintf(&b, "\tDeprecated: %s\n", info.DeprecationReason)
		} else {
			fmt.Fprintf(&b, "\tDeprecated: true\n")
		}
	}

	if len(info.Args) > 0 {
		fmt.Fprintf(&b, "\tArgs:\n")
		for _, arg := range info.Args {
			fmt.Fprintf(&b, "\t\t%s\n", formatArg(arg))
		}
	}

	return b.String()
}

func formatArg(arg ArgInfo) string {
	summary := fmt.Sprintf("%s: %s", arg.Name, arg.Type)

	if arg.Description != "" {
		summary += " - " + arg.Description
	}

	return summary
}
