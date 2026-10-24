package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/viper"
)

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
