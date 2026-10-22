package grpcclient

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func FormatMethodSummary(info *MethodInfo) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n", info.QualifiedName())

	if info.IsClientStreaming || info.IsServerStreaming {
		fmt.Fprintf(&b, "\tStreaming: client=%t, server=%t\n", info.IsClientStreaming, info.IsServerStreaming)
	}

	fmt.Fprintf(&b, "\tRequest: %s\n", info.Input.FullName())
	formatMessageFields(&b, info.Input, "\t\t", map[protoreflect.FullName]bool{})

	fmt.Fprintf(&b, "\tResponse: %s\n", info.Output.FullName())
	formatMessageFields(&b, info.Output, "\t\t", map[protoreflect.FullName]bool{})

	return b.String()
}

func formatMessageFields(b *strings.Builder, msg protoreflect.MessageDescriptor, indent string, seen map[protoreflect.FullName]bool) {
	if seen[msg.FullName()] {
		fmt.Fprintf(b, "%s... (recursive reference to %s, see above)\n", indent, msg.FullName())
		return
	}

	fields := msg.Fields()
	if fields.Len() == 0 {
		fmt.Fprintf(b, "%s(no fields)\n", indent)
		return
	}

	seen[msg.FullName()] = true
	defer delete(seen, msg.FullName())

	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fmt.Fprintf(b, "%s%s\n", indent, formatField(f))

		switch {
		case f.IsMap():
			if valueField := f.MapValue(); valueField.Kind() == protoreflect.MessageKind {
				formatMessageFields(b, valueField.Message(), indent+"\t", seen)
			}
		case f.Kind() == protoreflect.MessageKind || f.Kind() == protoreflect.GroupKind:
			formatMessageFields(b, f.Message(), indent+"\t", seen)
		case f.Kind() == protoreflect.EnumKind:
			formatEnumValues(b, f.Enum(), indent+"\t")
		}
	}
}

func formatField(f protoreflect.FieldDescriptor) string {
	var attrs []string

	switch {
	case f.IsMap():
		attrs = append(attrs, "map")
	case f.IsList():
		attrs = append(attrs, "repeated")
	default:
		attrs = append(attrs, "singular")
	}

	if oneof := f.ContainingOneof(); oneof != nil && !oneof.IsSynthetic() {
		attrs = append(attrs, fmt.Sprintf("oneof %s", oneof.Name()))
	} else if f.HasOptionalKeyword() {
		attrs = append(attrs, "explicit presence")
	}

	return fmt.Sprintf("%s %s (%s, #%d)", f.Name(), fieldTypeName(f), strings.Join(attrs, ", "), f.Number())
}

func fieldTypeName(f protoreflect.FieldDescriptor) string {
	switch f.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if f.IsMap() {
			return fmt.Sprintf("map<%s, %s>", f.MapKey().Kind(), fieldTypeName(f.MapValue()))
		}
		return string(f.Message().FullName())
	case protoreflect.EnumKind:
		return string(f.Enum().FullName())
	default:
		return f.Kind().String()
	}
}

func formatEnumValues(b *strings.Builder, enum protoreflect.EnumDescriptor, indent string) {
	values := enum.Values()

	names := make([]string, values.Len())
	for i := 0; i < values.Len(); i++ {
		names[i] = string(values.Get(i).Name())
	}

	fmt.Fprintf(b, "%svalues: %s\n", indent, strings.Join(names, ", "))
}
