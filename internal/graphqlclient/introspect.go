// Package graphqlclient discovers GraphQL operations via introspection, the
// GraphQL analogue of OAS specs and gRPC reflection elsewhere in apix.
package graphqlclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/httpclient"
)

// TypeRef only unwraps 4 levels of LIST/NON_NULL (e.g. "[[String!]!]!"),
// enough for real-world schemas without a full introspection dump's weight.
const introspectionQuery = `query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      kind
      name
      fields(includeDeprecated: true) {
        name
        description
        args {
          name
          description
          type { ...TypeRef }
          defaultValue
        }
        type { ...TypeRef }
        isDeprecated
        deprecationReason
      }
      inputFields {
        name
        description
        type { ...TypeRef }
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
    }
  }
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
        }
      }
    }
  }
}`

// StatusError marks a non-2xx introspection response, so callers can
// special-case e.g. 401s like the HTTP and gRPC handlers do.
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("introspection query returned status %d", e.StatusCode)
}

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   string   `json:"name"`
	OfType *typeRef `json:"ofType"`
}

func (t *typeRef) String() string {
	if t == nil {
		return "Unknown"
	}

	switch t.Kind {
	case "NON_NULL":
		return t.OfType.String() + "!"
	case "LIST":
		return "[" + t.OfType.String() + "]"
	default:
		return t.Name
	}
}

type inputValue struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         typeRef `json:"type"`
	DefaultValue *string `json:"defaultValue"`
}

type schemaField struct {
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Args              []inputValue `json:"args"`
	Type              typeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

type enumValueDef struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type fullType struct {
	Kind        string         `json:"kind"`
	Name        string         `json:"name"`
	Fields      []schemaField  `json:"fields"`
	InputFields []inputValue   `json:"inputFields"`
	EnumValues  []enumValueDef `json:"enumValues"`
}

type namedType struct {
	Name string `json:"name"`
}

type schema struct {
	QueryType        *namedType `json:"queryType"`
	MutationType     *namedType `json:"mutationType"`
	SubscriptionType *namedType `json:"subscriptionType"`
	Types            []fullType `json:"types"`
}

type introspectionResponse struct {
	Data struct {
		Schema schema `json:"__schema"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type ArgInfo struct {
	Name        string
	Type        string
	Description string

	InputFields []ArgInfo // set if this arg's type is an input object
	EnumValues  []string  // set if this arg's type is an enum

	IsRecursiveRef bool // set instead of expanding InputFields into a cycle
}

type FieldInfo struct {
	Name              string
	Type              string
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	IsExpandable      bool
	HasRequiredArgs   bool
}

type OperationInfo struct {
	Kind              string // "Query", "Mutation", or "Subscription"
	Name              string
	Args              []ArgInfo
	ReturnType        string
	Description       string
	IsDeprecated      bool
	DeprecationReason string
	Fields            []FieldInfo // fields selectable on ReturnType, one level deep; nil for scalar/enum returns
}

func (o OperationInfo) QualifiedName() string {
	return o.Kind + "." + o.Name
}

// Signature renders e.g. "Query.user(id: ID!) -> User".
func (o OperationInfo) Signature() string {
	args := make([]string, len(o.Args))
	for i, a := range o.Args {
		args[i] = a.Name + ": " + a.Type
	}

	return fmt.Sprintf("%s(%s) -> %s", o.QualifiedName(), strings.Join(args, ", "), o.ReturnType)
}

func Introspect(domain *config.Domain, headers map[string]string) (*schema, error) {
	body, err := json.Marshal(map[string]string{"query": introspectionQuery})
	if err != nil {
		return nil, fmt.Errorf("failed to build introspection query: %w", err)
	}

	resp, err := httpclient.Do(http.MethodPost, domain, "", &body, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to query introspection endpoint: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read introspection response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}

	var parsed introspectionResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("introspection query failed: %s", parsed.Errors[0].Message)
	}

	if parsed.Data.Schema.QueryType == nil {
		return nil, fmt.Errorf("introspection response did not include a schema")
	}

	return &parsed.Data.Schema, nil
}

// ListOperations groups Query, Mutation, then Subscription fields, sorted by
// name within each group.
func ListOperations(domain *config.Domain, headers map[string]string) ([]OperationInfo, error) {
	s, err := Introspect(domain, headers)
	if err != nil {
		return nil, err
	}

	var infos []OperationInfo

	for _, root := range rootKinds(s) {
		fields := findFields(s, root.typeName)

		ops := make([]OperationInfo, 0, len(fields))
		for _, f := range fields {
			ops = append(ops, toOperationInfo(s, root.kind, f))
		}

		sort.Slice(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
		infos = append(infos, ops...)
	}

	return infos, nil
}

// ResolveOperation accepts a bare name (e.g. "user") or one qualified by
// root type (e.g. "Query.user"), analogous to grpcclient.ResolveMethod.
func ResolveOperation(domain *config.Domain, name string, headers map[string]string) (*OperationInfo, error) {
	s, err := Introspect(domain, headers)
	if err != nil {
		return nil, err
	}

	match, path, err := resolveFieldPath(s, name)
	if err != nil {
		return nil, err
	}

	if len(path) == 0 {
		info := toOperationInfo(s, match.rootKind, match.field)
		return &info, nil
	}

	return descendFieldPath(s, match.rootKind, match.field.Name, match.field.Type, path)
}

type fieldMatch struct {
	rootKind string
	field    schemaField
}

type notFoundError struct {
	name string
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("no operation named %q found via introspection", e.name)
}

func resolveField(s *schema, name, suffix string) (*fieldMatch, error) {
	qualifiedKind, bareName, isQualified := splitQualified(name)

	var candidates []fieldMatch

	for _, root := range rootKinds(s) {
		for _, f := range findFields(s, root.typeName) {
			if isQualified {
				if root.kind == qualifiedKind && f.Name == bareName {
					candidates = append(candidates, fieldMatch{root.kind, f})
				}
				continue
			}

			if f.Name == name {
				candidates = append(candidates, fieldMatch{root.kind, f})
			}
		}
	}

	if len(candidates) == 0 {
		return nil, &notFoundError{name: name}
	}

	if len(candidates) > 1 {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = toOperationInfo(s, c.rootKind, c.field).QualifiedName()
		}
		sort.Strings(names)

		return nil, fmt.Errorf(
			"operation name %q is ambiguous, found on multiple root types: %s\nqualify it, e.g. `apix show %s`",
			name, strings.Join(names, ", "), names[0]+suffix,
		)
	}

	return &candidates[0], nil
}

func resolveFieldPath(s *schema, name string) (*fieldMatch, []string, error) {
	segments := strings.Split(name, ".")

	end := 1
	if len(segments) > 1 {
		switch segments[0] {
		case "Query", "Mutation", "Subscription":
			end = 2
		}
	}

	suffix := ""
	if end < len(segments) {
		suffix = "." + strings.Join(segments[end:], ".")
	}
	match, err := resolveField(s, strings.Join(segments[:end], "."), suffix)
	if err != nil {
		return nil, nil, err
	}
	return match, segments[end:], nil
}

// descendFieldPath walks path, a sequence of field names, from fieldType -
// the return type of the baseName field already resolved against rootKind -
// and describes the field found at the end of it as its own OperationInfo.
func descendFieldPath(s *schema, rootKind, baseName string, fieldType typeRef, path []string) (*OperationInfo, error) {
	cur := fieldType
	name := baseName

	var f schemaField

	for _, seg := range path {
		t := baseType(cur)

		if t.Kind != "OBJECT" && t.Kind != "INTERFACE" {
			return nil, fmt.Errorf("field %q is not an object and cannot be traversed further (in path %q)", name, strings.Join(path, "."))
		}

		found := false
		for _, cf := range findFields(s, t.Name) {
			if cf.Name == seg {
				f, found = cf, true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("unknown field %q on type %q", seg, t.Name)
		}

		name += "." + seg
		cur = f.Type
	}

	args := make([]ArgInfo, len(f.Args))
	for i, a := range f.Args {
		args[i] = toArgInfo(s, a, map[string]bool{})
	}

	return &OperationInfo{
		Kind:              rootKind,
		Name:              name,
		Args:              args,
		ReturnType:        f.Type.String(),
		Description:       f.Description,
		IsDeprecated:      f.IsDeprecated,
		DeprecationReason: f.DeprecationReason,
		Fields:            resultFields(s, f.Type),
	}, nil
}

type rootKind struct {
	kind     string
	typeName string
}

func rootKinds(s *schema) []rootKind {
	var roots []rootKind

	if s.QueryType != nil {
		roots = append(roots, rootKind{"Query", s.QueryType.Name})
	}
	if s.MutationType != nil {
		roots = append(roots, rootKind{"Mutation", s.MutationType.Name})
	}
	if s.SubscriptionType != nil {
		roots = append(roots, rootKind{"Subscription", s.SubscriptionType.Name})
	}

	return roots
}

func findType(s *schema, typeName string) *fullType {
	for i := range s.Types {
		if s.Types[i].Name == typeName {
			return &s.Types[i]
		}
	}

	return nil
}

func findFields(s *schema, typeName string) []schemaField {
	if t := findType(s, typeName); t != nil {
		return t.Fields
	}

	return nil
}

func toOperationInfo(s *schema, kind string, f schemaField) OperationInfo {
	args := make([]ArgInfo, len(f.Args))
	for i, a := range f.Args {
		args[i] = toArgInfo(s, a, map[string]bool{})
	}

	return OperationInfo{
		Kind:              kind,
		Name:              f.Name,
		Args:              args,
		ReturnType:        f.Type.String(),
		Description:       f.Description,
		IsDeprecated:      f.IsDeprecated,
		DeprecationReason: f.DeprecationReason,
		Fields:            resultFields(s, f.Type),
	}
}

func resultFields(s *schema, t typeRef) []FieldInfo {
	base := baseType(t)
	if base.Kind != "OBJECT" && base.Kind != "INTERFACE" {
		return nil
	}

	fields := findFields(s, base.Name)
	infos := make([]FieldInfo, len(fields))

	for i, f := range fields {
		fb := baseType(f.Type)

		infos[i] = FieldInfo{
			Name:              f.Name,
			Type:              f.Type.String(),
			Description:       f.Description,
			IsDeprecated:      f.IsDeprecated,
			DeprecationReason: f.DeprecationReason,
			IsExpandable:      fb.Kind == "OBJECT" || fb.Kind == "INTERFACE",
			HasRequiredArgs:   hasRequiredArg(f.Args),
		}
	}

	return infos
}

func hasRequiredArg(args []inputValue) bool {
	for _, a := range args {
		if a.Type.Kind == "NON_NULL" && a.DefaultValue == nil {
			return true
		}
	}

	return false
}

func toArgInfo(s *schema, a inputValue, seen map[string]bool) ArgInfo {
	info := ArgInfo{
		Name:        a.Name,
		Type:        a.Type.String(),
		Description: a.Description,
	}

	t := findType(s, baseType(a.Type).Name)
	if t == nil {
		return info
	}

	switch t.Kind {
	case "INPUT_OBJECT":
		if seen[t.Name] {
			info.IsRecursiveRef = true
			return info
		}

		seen[t.Name] = true
		defer delete(seen, t.Name)

		info.InputFields = make([]ArgInfo, len(t.InputFields))
		for i, f := range t.InputFields {
			info.InputFields[i] = toArgInfo(s, f, seen)
		}
	case "ENUM":
		for _, v := range t.EnumValues {
			info.EnumValues = append(info.EnumValues, v.Name)
		}
	}

	return info
}

func splitQualified(name string) (kind, bareName string, ok bool) {
	idx := strings.Index(name, ".")
	if idx < 0 {
		return "", "", false
	}

	return name[:idx], name[idx+1:], true
}
