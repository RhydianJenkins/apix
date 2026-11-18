package graphqlclient_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
)

const testSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": {"name": "Mutation"},
      "subscriptionType": null,
      "types": [
        {
          "kind": "OBJECT",
          "name": "Query",
          "fields": [
            {
              "name": "user",
              "description": "Fetch a user by ID.",
              "args": [
                {"name": "id", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}}
              ],
              "type": {"kind": "OBJECT", "name": "User", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "search",
              "description": "",
              "args": [
                {"name": "term", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}}
              ],
              "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "LIST", "name": null, "ofType": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}}},
              "isDeprecated": true,
              "deprecationReason": "Use user search instead."
            }
          ]
        },
        {
          "kind": "OBJECT",
          "name": "Mutation",
          "fields": [
            {
              "name": "createUser",
              "description": "Create a new user.",
              "args": [
                {"name": "input", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "INPUT_OBJECT", "name": "CreateUserInput", "ofType": null}}}
              ],
              "type": {"kind": "OBJECT", "name": "User", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ]
        }
      ]
    }
  }
}`

func newTestDomain(t *testing.T, handler http.HandlerFunc) *config.Domain {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	}
}

func schemaHandler(t *testing.T) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testSchemaResponse))
	}
}

func TestListOperations(t *testing.T) {
	domain := newTestDomain(t, schemaHandler(t))

	infos, err := graphqlclient.ListOperations(domain, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(infos) != 3 {
		t.Fatalf("expected 3 operations, got %d: %+v", len(infos), infos)
	}

	// Query operations come first, sorted by name ("search" < "user"),
	// followed by Mutation operations.
	want := []string{"Query.search", "Query.user", "Mutation.createUser"}
	for i, w := range want {
		if infos[i].QualifiedName() != w {
			t.Errorf("expected operation %d to be %q, got %q", i, w, infos[i].QualifiedName())
		}
	}

	user := infos[1]
	if user.ReturnType != "User" {
		t.Errorf("expected user.ReturnType to be %q, got %q", "User", user.ReturnType)
	}
	if len(user.Args) != 1 || user.Args[0].Name != "id" || user.Args[0].Type != "ID!" {
		t.Errorf("expected user to have a single id: ID! arg, got %+v", user.Args)
	}

	search := infos[0]
	if search.ReturnType != "[String!]!" {
		t.Errorf("expected search.ReturnType to be %q, got %q", "[String!]!", search.ReturnType)
	}
	if !search.IsDeprecated || search.DeprecationReason != "Use user search instead." {
		t.Errorf("expected search to be deprecated with a reason, got %+v", search)
	}
}

func TestResolveOperation_Bare(t *testing.T) {
	domain := newTestDomain(t, schemaHandler(t))

	info, err := graphqlclient.ResolveOperation(domain, "user", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.QualifiedName() != "Query.user" {
		t.Errorf("expected Query.user, got %q", info.QualifiedName())
	}
}

func TestResolveOperation_Qualified(t *testing.T) {
	domain := newTestDomain(t, schemaHandler(t))

	info, err := graphqlclient.ResolveOperation(domain, "Mutation.createUser", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "createUser" || info.Kind != "Mutation" {
		t.Errorf("expected Mutation.createUser, got %+v", info)
	}
}

func TestResolveOperation_NotFound(t *testing.T) {
	domain := newTestDomain(t, schemaHandler(t))

	_, err := graphqlclient.ResolveOperation(domain, "doesNotExist", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown operation, got nil")
	}

	if !strings.Contains(err.Error(), "no operation named") {
		t.Errorf("expected a 'no operation named' error, got: %v", err)
	}
}

func TestIntrospect_SendsHeaders(t *testing.T) {
	var gotAuth string

	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testSchemaResponse))
	})

	_, err := graphqlclient.ListOperations(domain, map[string]string{"Authorization": "Bearer test-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected authorization header to reach the introspection call, got %q", gotAuth)
	}
}

func TestIntrospect_NonSuccessStatusReturnsStatusError(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	})

	_, err := graphqlclient.Introspect(domain, nil)
	if err == nil {
		t.Fatal("expected an error for a non-2xx response, got nil")
	}

	var statusErr *graphqlclient.StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected a *StatusError, got: %v", err)
	}

	if statusErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", statusErr.StatusCode)
	}
}

const expandedArgsSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": null,
      "subscriptionType": null,
      "types": [
        {
          "kind": "OBJECT",
          "name": "Query",
          "fields": [
            {
              "name": "account",
              "description": "",
              "args": [
                {"name": "identifier", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "INPUT_OBJECT", "name": "AccountIdentifier", "ofType": null}}},
                {"name": "state", "description": "", "type": {"kind": "ENUM", "name": "AccountState", "ofType": null}},
                {"name": "self", "description": "", "type": {"kind": "INPUT_OBJECT", "name": "SelfRefInput", "ofType": null}}
              ],
              "type": {"kind": "OBJECT", "name": "Account", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ],
          "inputFields": null,
          "enumValues": null
        },
        {
          "kind": "INPUT_OBJECT",
          "name": "AccountIdentifier",
          "fields": null,
          "inputFields": [
            {"name": "id", "description": "Account UUID.", "type": {"kind": "SCALAR", "name": "String", "ofType": null}},
            {"name": "number", "description": "", "type": {"kind": "SCALAR", "name": "String", "ofType": null}}
          ],
          "enumValues": null
        },
        {
          "kind": "ENUM",
          "name": "AccountState",
          "fields": null,
          "inputFields": null,
          "enumValues": [
            {"name": "ACTIVE", "description": "", "isDeprecated": false, "deprecationReason": null},
            {"name": "CLOSED", "description": "", "isDeprecated": false, "deprecationReason": null}
          ]
        },
        {
          "kind": "INPUT_OBJECT",
          "name": "SelfRefInput",
          "fields": null,
          "inputFields": [
            {"name": "value", "description": "", "type": {"kind": "SCALAR", "name": "String", "ofType": null}},
            {"name": "nested", "description": "", "type": {"kind": "INPUT_OBJECT", "name": "SelfRefInput", "ofType": null}}
          ],
          "enumValues": null
        }
      ]
    }
  }
}`

func TestResolveOperation_ExpandsInputObjectArgs(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(expandedArgsSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var identifier *graphqlclient.ArgInfo
	for i := range info.Args {
		if info.Args[i].Name == "identifier" {
			identifier = &info.Args[i]
		}
	}

	if identifier == nil {
		t.Fatalf("expected an 'identifier' arg, got %+v", info.Args)
	}

	if len(identifier.InputFields) != 2 {
		t.Fatalf("expected 2 input fields, got %d: %+v", len(identifier.InputFields), identifier.InputFields)
	}

	if identifier.InputFields[0].Name != "id" || identifier.InputFields[0].Description != "Account UUID." {
		t.Errorf("expected the id field with its description, got %+v", identifier.InputFields[0])
	}

	if identifier.InputFields[1].Name != "number" {
		t.Errorf("expected the number field, got %+v", identifier.InputFields[1])
	}
}

func TestResolveOperation_ExpandsEnumArgValues(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(expandedArgsSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var state *graphqlclient.ArgInfo
	for i := range info.Args {
		if info.Args[i].Name == "state" {
			state = &info.Args[i]
		}
	}

	if state == nil {
		t.Fatalf("expected a 'state' arg, got %+v", info.Args)
	}

	want := []string{"ACTIVE", "CLOSED"}
	if !slices.Equal(state.EnumValues, want) {
		t.Errorf("expected enum values %v, got %v", want, state.EnumValues)
	}
}

func TestResolveOperation_GuardsAgainstSelfReferencingInput(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(expandedArgsSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var self *graphqlclient.ArgInfo
	for i := range info.Args {
		if info.Args[i].Name == "self" {
			self = &info.Args[i]
		}
	}

	if self == nil {
		t.Fatalf("expected a 'self' arg, got %+v", info.Args)
	}

	var nested *graphqlclient.ArgInfo
	for i := range self.InputFields {
		if self.InputFields[i].Name == "nested" {
			nested = &self.InputFields[i]
		}
	}

	if nested == nil {
		t.Fatalf("expected a 'nested' field, got %+v", self.InputFields)
	}

	if !nested.IsRecursiveRef {
		t.Errorf("expected the self-referencing field to be flagged recursive, got %+v", nested)
	}

	if len(nested.InputFields) != 0 {
		t.Errorf("expected recursion to stop instead of expanding forever, got %+v", nested.InputFields)
	}
}

const fieldPathSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": null,
      "subscriptionType": null,
      "types": [
        {
          "kind": "OBJECT",
          "name": "Query",
          "fields": [
            {
              "name": "account",
              "description": "",
              "args": [],
              "type": {"kind": "OBJECT", "name": "Account", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ]
        },
        {
          "kind": "OBJECT",
          "name": "Account",
          "fields": [
            {
              "name": "number",
              "description": "The account number.",
              "args": [],
              "type": {"kind": "SCALAR", "name": "String", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "serviceTransfers",
              "description": "",
              "args": [],
              "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "LIST", "name": null, "ofType": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "OBJECT", "name": "ServiceTransfer", "ofType": null}}}},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ]
        },
        {
          "kind": "OBJECT",
          "name": "ServiceTransfer",
          "fields": [
            {
              "name": "id",
              "description": "The transfer's ID.",
              "args": [],
              "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "status",
              "description": "",
              "args": [],
              "type": {"kind": "ENUM", "name": "TransferStatus", "ofType": null},
              "isDeprecated": true,
              "deprecationReason": "Use state instead."
            }
          ]
        }
      ]
    }
  }
}`

func TestResolveOperation_ListsResultFields(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fieldPathSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "Query.account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(info.Fields) != 2 {
		t.Fatalf("expected 2 result fields, got %d: %+v", len(info.Fields), info.Fields)
	}

	if info.Fields[0].Name != "number" || info.Fields[0].IsExpandable {
		t.Errorf("expected a non-expandable scalar 'number' field, got %+v", info.Fields[0])
	}

	if info.Fields[1].Name != "serviceTransfers" || !info.Fields[1].IsExpandable {
		t.Errorf("expected an expandable 'serviceTransfers' field, got %+v", info.Fields[1])
	}
}

const requiredArgSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": null,
      "subscriptionType": null,
      "types": [
        {
          "kind": "OBJECT",
          "name": "Query",
          "fields": [
            {
              "name": "account",
              "description": "",
              "args": [],
              "type": {"kind": "OBJECT", "name": "Account", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ]
        },
        {
          "kind": "OBJECT",
          "name": "Account",
          "fields": [
            {
              "name": "number",
              "description": "",
              "args": [],
              "type": {"kind": "SCALAR", "name": "String", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "balanceAt",
              "description": "",
              "args": [
                {"name": "date", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}, "defaultValue": null}
              ],
              "type": {"kind": "SCALAR", "name": "Float", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "balanceWithFallback",
              "description": "",
              "args": [
                {"name": "currency", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}, "defaultValue": "\"GBP\""}
              ],
              "type": {"kind": "SCALAR", "name": "Float", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            },
            {
              "name": "nickname",
              "description": "",
              "args": [
                {"name": "locale", "description": "", "type": {"kind": "SCALAR", "name": "String", "ofType": null}, "defaultValue": null}
              ],
              "type": {"kind": "SCALAR", "name": "String", "ofType": null},
              "isDeprecated": false,
              "deprecationReason": null
            }
          ]
        }
      ]
    }
  }
}`

// TestResolveOperation_FieldsWithRequiredArgs locks in which fields count as
// having a "required" argument for FieldInfo.HasRequiredArgs: non-null with
// no default (balanceAt) is required; non-null with a default
// (balanceWithFallback) and nullable without a default (nickname) are not.
func TestResolveOperation_FieldsWithRequiredArgs(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(requiredArgSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "Query.account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	byName := make(map[string]graphqlclient.FieldInfo, len(info.Fields))
	for _, f := range info.Fields {
		byName[f.Name] = f
	}

	cases := []struct {
		name string
		want bool
	}{
		{"number", false},
		{"balanceAt", true},
		{"balanceWithFallback", false},
		{"nickname", false},
	}

	for _, c := range cases {
		f, ok := byName[c.name]
		if !ok {
			t.Fatalf("expected a %q field, got %+v", c.name, info.Fields)
		}

		if f.HasRequiredArgs != c.want {
			t.Errorf("field %q: expected HasRequiredArgs=%v, got %v", c.name, c.want, f.HasRequiredArgs)
		}
	}
}

func TestResolveOperation_FieldPath_Qualified(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fieldPathSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "Query.account.serviceTransfers", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.QualifiedName() != "Query.account.serviceTransfers" {
		t.Errorf("expected QualifiedName %q, got %q", "Query.account.serviceTransfers", info.QualifiedName())
	}

	if info.ReturnType != "[ServiceTransfer!]!" {
		t.Errorf("expected ReturnType %q, got %q", "[ServiceTransfer!]!", info.ReturnType)
	}

	if len(info.Fields) != 2 {
		t.Fatalf("expected 2 result fields, got %d: %+v", len(info.Fields), info.Fields)
	}

	if info.Fields[0].Name != "id" || info.Fields[1].Name != "status" {
		t.Errorf("expected id/status fields on ServiceTransfer, got %+v", info.Fields)
	}

	if !info.Fields[1].IsDeprecated || info.Fields[1].DeprecationReason != "Use state instead." {
		t.Errorf("expected status to be deprecated with a reason, got %+v", info.Fields[1])
	}
}

func TestResolveOperation_FieldPath_Unqualified(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fieldPathSchemaResponse))
	})

	info, err := graphqlclient.ResolveOperation(domain, "account.serviceTransfers", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.QualifiedName() != "Query.account.serviceTransfers" {
		t.Errorf("expected QualifiedName %q, got %q", "Query.account.serviceTransfers", info.QualifiedName())
	}
}

func TestResolveOperation_FieldPath_UnknownField(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fieldPathSchemaResponse))
	})

	_, err := graphqlclient.ResolveOperation(domain, "Query.account.doesNotExist", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown nested field, got nil")
	}

	if !strings.Contains(err.Error(), `unknown field "doesNotExist"`) {
		t.Errorf("expected an 'unknown field' error, got: %v", err)
	}
}

func TestResolveOperation_FieldPath_NotTraversable(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fieldPathSchemaResponse))
	})

	_, err := graphqlclient.ResolveOperation(domain, "Query.account.number.whatever", nil)
	if err == nil {
		t.Fatal("expected an error when traversing past a scalar field, got nil")
	}

	if !strings.Contains(err.Error(), "is not an object and cannot be traversed further") {
		t.Errorf("expected a 'not traversable' error, got: %v", err)
	}
}

func TestIntrospect_GraphQLErrorsAreSurfaced(t *testing.T) {
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors": [{"message": "introspection is disabled"}]}`))
	})

	_, err := graphqlclient.Introspect(domain, nil)
	if err == nil {
		t.Fatal("expected an error when the response contains GraphQL errors, got nil")
	}

	if !strings.Contains(err.Error(), "introspection is disabled") {
		t.Errorf("expected the GraphQL error message to be surfaced, got: %v", err)
	}
}

func TestResolveOperation_FieldPath_RootQualifiers(t *testing.T) {
	const response = `{"data":{"__schema":{
		"queryType":{"name":"QueryRoot"},
		"mutationType":{"name":"MutationRoot"},
		"types":[
			{"name":"QueryRoot","fields":[
				{"name":"Query","type":{"kind":"OBJECT","name":"Container"}},
				{"name":"Mutation","type":{"kind":"OBJECT","name":"Container"}},
				{"name":"Subscription","type":{"kind":"OBJECT","name":"Container"}},
				{"name":"account","type":{"kind":"OBJECT","name":"Account"}}
			]},
			{"name":"MutationRoot","fields":[
				{"name":"account","type":{"kind":"OBJECT","name":"Account"}}
			]},
			{"name":"Container","fields":[
				{"name":"account","type":{"kind":"OBJECT","name":"Account"}},
				{"name":"missing","type":{"kind":"SCALAR","name":"String"}}
			]},
			{"name":"Account","fields":[
				{"name":"number","type":{"kind":"SCALAR","name":"String"}}
			]}
		]
	}}}`
	domain := newTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	})

	for _, path := range []string{"Query.missing", "Mutation.missing", "Subscription.missing", "Subscription.account"} {
		t.Run(path, func(t *testing.T) {
			_, err := graphqlclient.ResolveOperation(domain, path, nil)
			if err == nil || !strings.Contains(err.Error(), "no operation named") {
				t.Fatalf("expected not found for explicit root %q, got %v", path, err)
			}
		})
	}
	info, err := graphqlclient.ResolveOperation(domain, "Query.Mutation.account.number", nil)
	if err != nil {
		t.Fatal(err)
	}
	if info.QualifiedName() != "Query.Mutation.account.number" || info.ReturnType != "String" {
		t.Fatalf("unexpected nested operation: %+v", info)
	}
	_, err = graphqlclient.ResolveOperation(domain, "account.number", nil)
	if err == nil || !strings.Contains(err.Error(), "`apix show Mutation.account.number`") {
		t.Fatalf("expected ambiguity hint to preserve nested path, got %v", err)
	}
}
