package graphqlclient_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
