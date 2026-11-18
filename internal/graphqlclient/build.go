package graphqlclient

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

type QueryDocument struct {
	Operation OperationInfo
	Query     string
}

// fieldPaths are dotted paths into nested object fields, relative to the
// operation's return type, extending the default selection (scalar/enum
// fields one level deep, falling back to "__typename" alone), e.g.
// "account.energy.properties".
func BuildQueryDocument(domain *config.Domain, name string, fieldPaths []string, headers map[string]string) (*QueryDocument, error) {
	s, err := Introspect(domain, headers)
	if err != nil {
		return nil, err
	}

	match, err := resolveField(s, name, "")
	if err != nil {
		return nil, err
	}

	if match.rootKind == "Subscription" {
		return nil, fmt.Errorf("%q is a Subscription; apix cannot invoke it over a plain HTTP POST (no websocket support)", name)
	}

	selection, err := buildSelection(s, match.field.Type, fieldPaths)
	if err != nil {
		return nil, err
	}

	return &QueryDocument{
		Operation: toOperationInfo(s, match.rootKind, match.field),
		Query:     renderDocument(match.rootKind, match.field, selection),
	}, nil
}

type selectionNode struct {
	name     string
	children []*selectionNode
}

func (n *selectionNode) findOrAddChild(name string) *selectionNode {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}

	child := &selectionNode{name: name}
	n.children = append(n.children, child)

	return child
}

// A scalar/enum returnType needs no selection at all.
func buildSelection(s *schema, returnType typeRef, fieldPaths []string) ([]*selectionNode, error) {
	base := baseType(returnType)

	if base.Kind == "SCALAR" || base.Kind == "ENUM" {
		if len(fieldPaths) > 0 {
			return nil, fmt.Errorf("operation returns %q, which has no fields to select; --fields is not applicable", base.Name)
		}

		return nil, nil
	}

	root := &selectionNode{name: "<root>"}
	applyDefaultSelection(root, s, base)

	for _, path := range fieldPaths {
		if err := applyFieldPath(root, s, base, strings.Split(path, ".")); err != nil {
			return nil, err
		}
	}

	return root.children, nil
}

func applyDefaultSelection(node *selectionNode, s *schema, t typeRef) {
	node.findOrAddChild("__typename")

	var leafNames []string
	for _, f := range findFields(s, t.Name) {
		if fb := baseType(f.Type); (fb.Kind == "SCALAR" || fb.Kind == "ENUM") && !hasRequiredArg(f.Args) {
			leafNames = append(leafNames, f.Name)
		}
	}
	sort.Strings(leafNames)

	for _, name := range leafNames {
		node.findOrAddChild(name)
	}
}

// The final segment also gets the default selection applied, if it's itself
// an object/interface/union, so the path stays useful on its own.
func applyFieldPath(node *selectionNode, s *schema, t typeRef, segments []string) error {
	if len(segments) == 0 {
		return nil
	}

	seg := segments[0]

	var matched *schemaField
	for _, f := range findFields(s, t.Name) {
		if f.Name == seg {
			fCopy := f
			matched = &fCopy
			break
		}
	}

	if matched == nil {
		return fmt.Errorf("unknown field %q on type %q", seg, t.Name)
	}

	if hasRequiredArg(matched.Args) {
		return fmt.Errorf("field %q requires an argument apix cannot supply when building a selection (in path %q)", seg, strings.Join(segments, "."))
	}

	child := node.findOrAddChild(seg)
	fb := baseType(matched.Type)

	if len(segments) == 1 {
		if fb.Kind == "OBJECT" || fb.Kind == "INTERFACE" || fb.Kind == "UNION" {
			applyDefaultSelection(child, s, fb)
		}

		return nil
	}

	if fb.Kind != "OBJECT" && fb.Kind != "INTERFACE" {
		return fmt.Errorf("field %q is not an object and cannot be traversed further (in path %q)", seg, strings.Join(segments, "."))
	}

	return applyFieldPath(child, s, fb, segments[1:])
}

func renderDocument(rootKind string, field schemaField, selection []*selectionNode) string {
	var b strings.Builder

	argDecls := make([]string, len(field.Args))
	callArgs := make([]string, len(field.Args))
	for i, a := range field.Args {
		argDecls[i] = fmt.Sprintf("$%s: %s", a.Name, a.Type.String())
		callArgs[i] = fmt.Sprintf("%s: $%s", a.Name, a.Name)
	}

	fmt.Fprintf(&b, "%s %s", strings.ToLower(rootKind), field.Name)
	if len(argDecls) > 0 {
		fmt.Fprintf(&b, "(%s)", strings.Join(argDecls, ", "))
	}
	b.WriteString(" {\n")

	fmt.Fprintf(&b, "  %s", field.Name)
	if len(callArgs) > 0 {
		fmt.Fprintf(&b, "(%s)", strings.Join(callArgs, ", "))
	}

	if len(selection) > 0 {
		b.WriteString(" {\n")
		renderSelection(&b, selection, "    ")
		b.WriteString("  }\n")
	} else {
		b.WriteString("\n")
	}

	b.WriteString("}\n")

	return b.String()
}

func renderSelection(b *strings.Builder, nodes []*selectionNode, indent string) {
	for _, n := range nodes {
		if len(n.children) == 0 {
			fmt.Fprintf(b, "%s%s\n", indent, n.name)
			continue
		}

		fmt.Fprintf(b, "%s%s {\n", indent, n.name)
		renderSelection(b, n.children, indent+"  ")
		fmt.Fprintf(b, "%s}\n", indent)
	}
}

// baseType unwraps NON_NULL/LIST wrappers down to the leaf named type.
func baseType(t typeRef) typeRef {
	cur := t
	for (cur.Kind == "NON_NULL" || cur.Kind == "LIST") && cur.OfType != nil {
		cur = *cur.OfType
	}

	return cur
}
