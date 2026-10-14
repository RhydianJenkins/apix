package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/grpctest"
	"github.com/rhydianjenkins/apix/v3/pkg/config"
)

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

func TestDiscoverCommand_HTTPNoOASReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"discover"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when no OAS spec is connected, got nil")
	}
}

func TestDiscoverCommand_HTTPListsEndpoints(t *testing.T) {
	specPath := writeTestOAS(t)

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	var output string
	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"discover"})

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

func TestDiscoverCommand_HTTPFiltersByMethod(t *testing.T) {
	specPath := writeTestOAS(t)

	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"discover", "--method", "post"})

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

func TestDiscoverCommand_GRPCListsMethods(t *testing.T) {
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
	cmd.SetArgs([]string{"discover"})

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

func TestDiscoverCommand_GRPCDialFailureReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: "127.0.0.1",
		GRPC: &config.GRPCOptions{Insecure: true, Port: 1},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"discover"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when reflection is unreachable, got nil")
	}
}
