package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/grpctest"
)

const graphqlTestSchema = `{
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
            }
          ]
        },
        {
          "kind": "OBJECT",
          "name": "Mutation",
          "fields": [
            {
              "name": "createUser",
              "description": "",
              "args": [],
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

func graphqlTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(graphqlTestSchema))
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

func writeTestOAS(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "oas.yaml")

	yamlContent := `openapi: 3.1.0
paths:
  /users:
    get:
      summary: List users
    post:
      summary: Create a user
  /users/{id}:
    get:
      summary: Get a user
`

	if err := os.WriteFile(specPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test OAS spec: %v", err)
	}

	return specPath
}

func TestShowCommand_HTTPNoOASReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when no OAS spec is connected, got nil")
	}
}

func TestShowCommand_HTTPListsEndpoints(t *testing.T) {
	specPath := writeTestOAS(t)

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	var output string
	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show"})

	err := func() error {
		var runErr error
		output = captureStdout(t, func() {
			runErr = cmd.Execute()
		})
		return runErr
	}()

	if err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	for _, want := range []string{"GET     /users - List users", "POST    /users - Create a user", "GET     /users/{id} - Get a user"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestShowCommand_HTTPFiltersByMethod(t *testing.T) {
	specPath := writeTestOAS(t)

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show", "--method", "post"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	if !strings.Contains(output, "POST    /users - Create a user") {
		t.Errorf("expected output to contain the POST endpoint, got:\n%s", output)
	}

	if strings.Contains(output, "GET") {
		t.Errorf("expected GET endpoints to be filtered out, got:\n%s", output)
	}
}

func TestShowCommand_GRPCListsMethods(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: srv.Host(),
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	want := grpctest.ServiceName + "/" + grpctest.MethodName + "(testpkg.GetReplyRequest) -> testpkg.Reply"
	if !strings.Contains(output, want) {
		t.Errorf("expected output to contain %q, got:\n%s", want, output)
	}
}

func TestShowCommand_GRPCDialFailureReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: "127.0.0.1",
		GRPC: &config.GRPCOptions{Insecure: true, Port: 1},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when reflection is unreachable, got nil")
	}
}

func TestShowCommand_GraphQLListsOperations(t *testing.T) {
	server := httptest.NewServer(graphqlTestHandler())
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	for _, want := range []string{"Query.user(id: ID!) -> User", "Mutation.createUser() -> User"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}

func TestShowCommand_GraphQLShowsSingleOperation(t *testing.T) {
	server := httptest.NewServer(graphqlTestHandler())
	defer server.Close()

	setActiveDomain(t, &config.Domain{
		Name: "testgql",
		Base: server.URL,
		HTTP: &config.HTTPOptions{GraphQL: true},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show", "Query.user"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	if !strings.Contains(output, "Query.user(id: ID!) -> User") {
		t.Errorf("expected output to contain the operation summary, got:\n%s", output)
	}

	if !strings.Contains(output, "Description: Fetch a user by ID.") {
		t.Errorf("expected output to contain the description, got:\n%s", output)
	}
}

func TestShowCommand_HTTPShowsSingleEndpoint(t *testing.T) {
	specPath := writeTestOAS(t)

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"show", "/users/{id}"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("expected Execute() to return nil, got: %v", err)
		}
	})

	if !strings.Contains(output, "GET /users/{id}") {
		t.Errorf("expected output to contain the endpoint summary, got:\n%s", output)
	}
}
