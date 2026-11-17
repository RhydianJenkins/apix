// Package graphqlclient discovers operations on a GraphQL domain via the
// standard introspection query, the GraphQL analogue of OAS specs (HTTP) and
// server reflection (gRPC) elsewhere in apix.
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

// introspectionQuery is the standard GraphQL introspection query, trimmed to
// only what apix needs to describe Query/Mutation/Subscription fields: their
// arguments and return types. The TypeRef fragment unwraps up to four levels
// of LIST/NON_NULL wrapping (e.g. "[[String!]!]!"), which covers real-world
// schemas without the depth (and weight) of a full introspection dump.
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
        }
        type { ...TypeRef }
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

// StatusError is returned by Introspect when the GraphQL endpoint responds
// with a non-2xx HTTP status, so callers can special-case things like 401s
// the same way the HTTP and gRPC handlers do.
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

// String renders a typeRef as GraphQL SDL, e.g. "[String!]!".
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
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        typeRef `json:"type"`
}

type schemaField struct {
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Args              []inputValue `json:"args"`
	Type              typeRef      `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason"`
}

type fullType struct {
	Name   string        `json:"name"`
	Fields []schemaField `json:"fields"`
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

// ArgInfo describes one argument accepted by a GraphQL operation.
type ArgInfo struct {
	Name        string
	Type        string
	Description string
}

// OperationInfo describes a single field on the Query, Mutation, or
// Subscription root type, resolved via introspection.
type OperationInfo struct {
	Kind              string // "Query", "Mutation", or "Subscription"
	Name              string
	Args              []ArgInfo
	ReturnType        string
	Description       string
	IsDeprecated      bool
	DeprecationReason string
}

// QualifiedName identifies an operation unambiguously, e.g. "Query.user".
func (o OperationInfo) QualifiedName() string {
	return o.Kind + "." + o.Name
}

// Signature renders a short one-line summary, e.g. "Query.user(id: ID!) -> User".
func (o OperationInfo) Signature() string {
	args := make([]string, len(o.Args))
	for i, a := range o.Args {
		args[i] = a.Name + ": " + a.Type
	}

	return fmt.Sprintf("%s(%s) -> %s", o.QualifiedName(), strings.Join(args, ", "), o.ReturnType)
}

// Introspect queries a GraphQL domain for its schema via the standard
// introspection query.
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

// ListOperations returns every Query, Mutation, and Subscription field found
// via introspection, grouped in that order and sorted by name within each
// group.
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
			ops = append(ops, toOperationInfo(root.kind, f))
		}

		sort.Slice(ops, func(i, j int) bool { return ops[i].Name < ops[j].Name })
		infos = append(infos, ops...)
	}

	return infos, nil
}

// ResolveOperation finds a single named operation, either bare (e.g. "user")
// or qualified by root type (e.g. "Query.user"), analogous to
// grpcclient.ResolveMethod.
func ResolveOperation(domain *config.Domain, name string, headers map[string]string) (*OperationInfo, error) {
	infos, err := ListOperations(domain, headers)
	if err != nil {
		return nil, err
	}

	qualifiedKind, bareName, isQualified := splitQualified(name)

	var candidates []OperationInfo

	for _, info := range infos {
		if isQualified {
			if info.Kind == qualifiedKind && info.Name == bareName {
				candidates = append(candidates, info)
			}
			continue
		}

		if info.Name == name {
			candidates = append(candidates, info)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no operation named %q found via introspection", name)
	}

	if len(candidates) > 1 {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.QualifiedName()
		}
		sort.Strings(names)

		return nil, fmt.Errorf(
			"operation name %q is ambiguous, found on multiple root types: %s\nqualify it, e.g. `apix show %s`",
			name, strings.Join(names, ", "), names[0],
		)
	}

	return &candidates[0], nil
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

func findFields(s *schema, typeName string) []schemaField {
	for _, t := range s.Types {
		if t.Name == typeName {
			return t.Fields
		}
	}

	return nil
}

func toOperationInfo(kind string, f schemaField) OperationInfo {
	args := make([]ArgInfo, len(f.Args))
	for i, a := range f.Args {
		args[i] = ArgInfo{
			Name:        a.Name,
			Type:        a.Type.String(),
			Description: a.Description,
		}
	}

	return OperationInfo{
		Kind:              kind,
		Name:              f.Name,
		Args:              args,
		ReturnType:        f.Type.String(),
		Description:       f.Description,
		IsDeprecated:      f.IsDeprecated,
		DeprecationReason: f.DeprecationReason,
	}
}

func splitQualified(name string) (kind, bareName string, ok bool) {
	idx := strings.Index(name, ".")
	if idx < 0 {
		return "", "", false
	}

	return name[:idx], name[idx+1:], true
}
