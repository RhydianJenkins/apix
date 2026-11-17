package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/viper"
)

// withStdin simulates `echo '...' | apix ...` by swapping os.Stdin for fn's duration.
func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()

	tmp, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatalf("failed to create stdin fixture: %v", err)
	}

	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("failed to write stdin fixture: %v", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatalf("failed to rewind stdin fixture: %v", err)
	}

	old := os.Stdin
	os.Stdin = tmp

	defer func() {
		os.Stdin = old
		tmp.Close()
	}()

	fn()
}

func setActiveDomain(t *testing.T, domain *config.Domain) {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".apix.yaml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	config.CfgPaths = []string{path}
	viper.Reset()

	if err := config.SetDomain(domain); err != nil {
		t.Fatalf("failed to set active domain: %v", err)
	}
}

func TestHTTPCommand_NonSuccessStatusReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "boom"}`))
	}))
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: server.URL,
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"get", "/anything"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error for a non-2xx response, got nil")
	}
}

func TestHTTPCommand_SuccessReturnsNilError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: server.URL,
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"get", "/anything"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil for a 200 response, got: %v", err)
	}
}

func TestHTTPCommand_GraphQLPostInvokesOperation(t *testing.T) {
	var gotBody map[string]any

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			w.Write([]byte(graphqlTestSchema)) // the introspection call
			return
		}

		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Write([]byte(`{"data": {"user": {"id": "123"}}}`))
	}))
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	})

	withStdin(t, `{"id": "123"}`, func() {
		cmd := newRootCmd("test")
		cmd.SetArgs([]string{"post", "Query.user"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	query, _ := gotBody["query"].(string)
	if query == "" {
		t.Fatal("expected a generated query to be sent, got none")
	}

	variables, ok := gotBody["variables"].(map[string]any)
	if !ok || variables["id"] != "123" {
		t.Errorf("expected variables id=123 from stdin, got %v", gotBody["variables"])
	}
}

func TestHTTPCommand_GraphQLPostUnknownOperationReturnsError(t *testing.T) {
	server := httptest.NewServer(graphqlTestHandler())
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	})

	withStdin(t, `{}`, func() {
		cmd := newRootCmd("test")
		cmd.SetArgs([]string{"post", "doesNotExist"})

		if err := cmd.Execute(); err == nil {
			t.Fatal("expected Execute() to return an error for an unknown operation, got nil")
		}
	})
}

func TestHTTPCommand_GraphQLPostSlashPathStaysRawPassthrough(t *testing.T) {
	var callCount int
	var gotRawBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		gotRawBody = string(body)
		w.Write([]byte(`{"data": {}}`))
	}))
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	})

	rawBody := `{"query": "query { __typename }"}`

	withStdin(t, rawBody, func() {
		cmd := newRootCmd("test")
		cmd.SetArgs([]string{"post", "/"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	// A single raw call, no introspection call first - confirms "/" bypasses
	// operation resolution entirely.
	if callCount != 1 {
		t.Fatalf("expected exactly 1 request (no introspection), got %d", callCount)
	}

	if gotRawBody != rawBody {
		t.Errorf("expected the raw body to be sent unchanged, got %q", gotRawBody)
	}
}

func TestHTTPCommand_WrongProtocolReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: "localhost",
		GRPC: &config.GRPCOptions{Insecure: true, Port: 50051},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"get", "/anything"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when the active domain is grpc, got nil")
	}
}
