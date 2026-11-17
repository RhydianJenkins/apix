package graphqlclient_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
)

const buildTestSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": {"name": "Mutation"},
      "subscriptionType": {"name": "Subscription"},
      "types": [
        {
          "name": "Query",
          "fields": [
            {
              "name": "user",
              "args": [{"name": "id", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}}],
              "type": {"kind": "OBJECT", "name": "User", "ofType": null}
            },
            {
              "name": "billUrl",
              "args": [{"name": "billId", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}}],
              "type": {"kind": "SCALAR", "name": "String", "ofType": null}
            }
          ]
        },
        {
          "name": "Mutation",
          "fields": [
            {
              "name": "createUser",
              "args": [{"name": "input", "description": "", "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "INPUT_OBJECT", "name": "CreateUserInput", "ofType": null}}}],
              "type": {"kind": "OBJECT", "name": "User", "ofType": null}
            }
          ]
        },
        {
          "name": "Subscription",
          "fields": [
            {
              "name": "notifications",
              "args": [],
              "type": {"kind": "OBJECT", "name": "Notification", "ofType": null}
            }
          ]
        },
        {
          "name": "User",
          "fields": [
            {"name": "id", "args": [], "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "ID", "ofType": null}}},
            {"name": "name", "args": [], "type": {"kind": "SCALAR", "name": "String", "ofType": null}},
            {"name": "tags", "args": [], "type": {"kind": "LIST", "name": null, "ofType": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}}},
            {"name": "address", "args": [], "type": {"kind": "OBJECT", "name": "Address", "ofType": null}}
          ]
        },
        {
          "name": "Address",
          "fields": [
            {"name": "street", "args": [], "type": {"kind": "SCALAR", "name": "String", "ofType": null}},
            {"name": "city", "args": [], "type": {"kind": "SCALAR", "name": "String", "ofType": null}}
          ]
        }
      ]
    }
  }
}`

func newBuildTestDomain(t *testing.T) *config.Domain {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(buildTestSchemaResponse))
	}))
	t.Cleanup(server.Close)

	return &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	}
}

func TestBuildQueryDocument_ObjectReturn_DefaultSelection(t *testing.T) {
	domain := newBuildTestDomain(t)

	doc, err := graphqlclient.BuildQueryDocument(domain, "user", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Operation.QualifiedName() != "Query.user" {
		t.Errorf("expected Query.user, got %q", doc.Operation.QualifiedName())
	}

	wantContains := []string{
		"query user($id: ID!) {",
		"user(id: $id) {",
		"__typename",
		"id",
		"name",
		"tags",
	}
	for _, want := range wantContains {
		if !strings.Contains(doc.Query, want) {
			t.Errorf("expected query to contain %q, got:\n%s", want, doc.Query)
		}
	}

	if strings.Contains(doc.Query, "address") {
		t.Errorf("expected the default selection to skip the object-typed address field, got:\n%s", doc.Query)
	}
}

func TestBuildQueryDocument_ScalarReturn_NoSelection(t *testing.T) {
	domain := newBuildTestDomain(t)

	doc, err := graphqlclient.BuildQueryDocument(domain, "billUrl", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(doc.Query, "{\n  billUrl") == false {
		t.Errorf("expected the query to call billUrl, got:\n%s", doc.Query)
	}

	if strings.Contains(doc.Query, "billUrl(billId: $billId) {") {
		t.Errorf("expected no selection set for a scalar return type, got:\n%s", doc.Query)
	}
}

func TestBuildQueryDocument_ScalarReturn_FieldsFlagRejected(t *testing.T) {
	domain := newBuildTestDomain(t)

	_, err := graphqlclient.BuildQueryDocument(domain, "billUrl", []string{"anything"}, nil)
	if err == nil {
		t.Fatal("expected an error when passing --fields for a scalar-returning operation, got nil")
	}

	if !strings.Contains(err.Error(), "no fields to select") {
		t.Errorf("expected a 'no fields to select' error, got: %v", err)
	}
}

func TestBuildQueryDocument_FieldsFlag_ExtendsSelection(t *testing.T) {
	domain := newBuildTestDomain(t)

	doc, err := graphqlclient.BuildQueryDocument(domain, "user", []string{"address.street"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(doc.Query, "address {") {
		t.Errorf("expected address to be selected, got:\n%s", doc.Query)
	}

	if !strings.Contains(doc.Query, "street") {
		t.Errorf("expected street to be selected within address, got:\n%s", doc.Query)
	}

	if strings.Contains(doc.Query, "city") {
		t.Errorf("expected city not to be selected (only the requested path), got:\n%s", doc.Query)
	}
}

func TestBuildQueryDocument_FieldsFlag_UnknownField(t *testing.T) {
	domain := newBuildTestDomain(t)

	_, err := graphqlclient.BuildQueryDocument(domain, "user", []string{"doesNotExist"}, nil)
	if err == nil {
		t.Fatal("expected an error for an unknown field path, got nil")
	}

	if !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected an 'unknown field' error, got: %v", err)
	}
}

func TestBuildQueryDocument_FieldsFlag_CannotTraverseScalar(t *testing.T) {
	domain := newBuildTestDomain(t)

	_, err := graphqlclient.BuildQueryDocument(domain, "user", []string{"name.sub"}, nil)
	if err == nil {
		t.Fatal("expected an error when traversing through a scalar field, got nil")
	}

	if !strings.Contains(err.Error(), "not an object") {
		t.Errorf("expected a 'not an object' error, got: %v", err)
	}
}

func TestBuildQueryDocument_Mutation(t *testing.T) {
	domain := newBuildTestDomain(t)

	doc, err := graphqlclient.BuildQueryDocument(domain, "Mutation.createUser", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(doc.Query, "mutation createUser($input: CreateUserInput!) {") {
		t.Errorf("expected a mutation document, got:\n%s", doc.Query)
	}
}

func TestBuildQueryDocument_RejectsSubscription(t *testing.T) {
	domain := newBuildTestDomain(t)

	_, err := graphqlclient.BuildQueryDocument(domain, "notifications", nil, nil)
	if err == nil {
		t.Fatal("expected an error for a Subscription, got nil")
	}

	if !strings.Contains(err.Error(), "Subscription") || !strings.Contains(err.Error(), "websocket") {
		t.Errorf("expected a helpful Subscription/websocket error, got: %v", err)
	}
}
