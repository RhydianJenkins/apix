package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

const graphqlShowTestSchema = `{
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
              "name": "user",
              "description": "Fetch a user by ID.",
              "args": [
                {"name": "id", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}}
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

func newGraphQLTestDomain(t *testing.T, handler http.HandlerFunc) *config.Domain {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	}
}

func TestGraphQLShowHandler(t *testing.T) {
	domain := newGraphQLTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(graphqlShowTestSchema))
	})

	summary, err := GraphQLShowHandler(domain, "user", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantContains := []string{
		"Query.user(id: ID!) -> User",
		"Description: Fetch a user by ID.",
		"id: ID!",
	}

	for _, want := range wantContains {
		if !strings.Contains(summary, want) {
			t.Errorf("expected summary to contain %q, got:\n%s", want, summary)
		}
	}
}

func TestGraphQLShowHandler_UnknownOperation(t *testing.T) {
	domain := newGraphQLTestDomain(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(graphqlShowTestSchema))
	})

	_, err := GraphQLShowHandler(domain, "doesNotExist", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown operation, got nil")
	}
}
