package graphqlclient

import (
	"fmt"
	"strings"
)

func FormatOperationSummary(info *OperationInfo, verbose bool) string {
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
			formatArg(&b, arg, "\t\t")
		}
	}

	if verbose && len(info.Fields) > 0 {
		fmt.Fprintf(&b, "\tFields:\n")
		for _, field := range info.Fields {
			formatField(&b, field, "\t\t")
		}
	}

	return b.String()
}

func formatField(b *strings.Builder, f FieldInfo, indent string) {
	summary := fmt.Sprintf("%s: %s", f.Name, f.Type)

	if f.Description != "" {
		summary += " - " + f.Description
	}

	fmt.Fprintf(b, "%s%s\n", indent, summary)

	if f.IsDeprecated {
		if f.DeprecationReason != "" {
			fmt.Fprintf(b, "%s\tDeprecated: %s\n", indent, f.DeprecationReason)
		} else {
			fmt.Fprintf(b, "%s\tDeprecated: true\n", indent)
		}
	}
}

func formatArg(b *strings.Builder, arg ArgInfo, indent string) {
	summary := fmt.Sprintf("%s: %s", arg.Name, arg.Type)

	if arg.Description != "" {
		summary += " - " + arg.Description
	}

	fmt.Fprintf(b, "%s%s\n", indent, summary)

	if arg.IsRecursiveRef {
		fmt.Fprintf(b, "%s\t... (recursive reference to %s, see above)\n", indent, arg.Type)
		return
	}

	if len(arg.EnumValues) > 0 {
		fmt.Fprintf(b, "%s\tvalues: %s\n", indent, strings.Join(arg.EnumValues, ", "))
	}

	if len(arg.InputFields) > 0 {
		fmt.Fprintf(b, "%s\tFields:\n", indent)
		for _, field := range arg.InputFields {
			formatArg(b, field, indent+"\t\t")
		}
	}
}
