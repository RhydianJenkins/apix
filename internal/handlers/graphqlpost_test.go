package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

func newGraphQLPostTestDomain(t *testing.T, onRequest func(body map[string]any)) *config.Domain {
	t.Helper()

	var callCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// The first request is always the introspection query.
			w.Write([]byte(graphqlShowTestSchema))
			return
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if onRequest != nil {
			onRequest(body)
		}

		w.Write([]byte(`{"data": {"user": {"id": "123"}}}`))
	}))
	t.Cleanup(server.Close)

	return &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	}
}

func TestGraphQLPostHandler(t *testing.T) {
	var gotBody map[string]any

	domain := newGraphQLPostTestDomain(t, func(body map[string]any) {
		gotBody = body
	})

	reqBody := []byte(`{"id": "123"}`)

	resp, err := GraphQLPostHandler(domain, "user", nil, &reqBody, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(resp), `"id": "123"`) {
		t.Errorf("expected the response body to be returned, got: %s", resp)
	}

	if gotBody["operationName"] != "user" {
		t.Errorf("expected operationName to be %q, got %v", "user", gotBody["operationName"])
	}

	query, _ := gotBody["query"].(string)
	if !strings.Contains(query, "query user($id: ID!) {") {
		t.Errorf("expected the built query to be sent, got: %s", query)
	}

	variables, ok := gotBody["variables"].(map[string]any)
	if !ok || variables["id"] != "123" {
		t.Errorf("expected variables to contain id=123, got %v", gotBody["variables"])
	}
}

func TestGraphQLPostHandler_NoStdin_SendsEmptyVariables(t *testing.T) {
	var gotBody map[string]any

	domain := newGraphQLPostTestDomain(t, func(body map[string]any) {
		gotBody = body
	})

	_, err := GraphQLPostHandler(domain, "user", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	variables, ok := gotBody["variables"].(map[string]any)
	if !ok || len(variables) != 0 {
		t.Errorf("expected empty variables, got %v", gotBody["variables"])
	}
}

func TestGraphQLPostHandler_InvalidStdinJSON(t *testing.T) {
	domain := newGraphQLPostTestDomain(t, nil)

	reqBody := []byte(`not json`)

	_, err := GraphQLPostHandler(domain, "user", nil, &reqBody, nil)
	if err == nil {
		t.Fatal("expected an error for invalid stdin JSON, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse stdin") {
		t.Errorf("expected a 'failed to parse stdin' error, got: %v", err)
	}
}

func TestGraphQLPostHandler_UnknownOperation(t *testing.T) {
	domain := newGraphQLPostTestDomain(t, nil)

	_, err := GraphQLPostHandler(domain, "doesNotExist", nil, nil, nil)
	if err == nil {
		t.Fatal("expected an error for an unknown operation, got nil")
	}

	if !strings.Contains(err.Error(), "failed to build query") {
		t.Errorf("expected a 'failed to build query' error, got: %v", err)
	}
}
