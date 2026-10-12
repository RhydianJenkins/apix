package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestDomain_IsGRPC(t *testing.T) {
	tests := map[string]struct {
		domain *Domain
		want   bool
	}{
		"nil domain":       {nil, false},
		"no options set":   {&Domain{}, false},
		"http options set": {&Domain{HTTP: &HTTPOptions{}}, false},
		"grpc options set": {&Domain{GRPC: &GRPCOptions{}}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.domain.IsGRPC(); got != tc.want {
				t.Errorf("IsGRPC() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSetAndLoadDomain_HTTP(t *testing.T) {
	tempDir := t.TempDir()
	CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	domain := &Domain{
		Name:    "http-api",
		Base:    "https://api.example.com",
		Headers: map[string]string{"X-Test": "value"},
		HTTP: &HTTPOptions{
			User:            "foo",
			Pass:            "bar",
			OpenAPISpecPath: "/tmp/spec.yaml",
			Port:            8443,
		},
	}

	SetDomain(domain)

	loaded, err := LoadDomain("http-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Base != domain.Base {
		t.Errorf("Base = %q, want %q", loaded.Base, domain.Base)
	}

	if loaded.IsGRPC() {
		t.Errorf("expected http domain, IsGRPC() returned true")
	}

	if loaded.HTTP == nil {
		t.Fatalf("expected HTTP options to round-trip, got nil")
	}

	if loaded.HTTP.User != "foo" || loaded.HTTP.Pass != "bar" || loaded.HTTP.OpenAPISpecPath != "/tmp/spec.yaml" || loaded.HTTP.Port != 8443 {
		t.Errorf("HTTP options did not round-trip correctly, got: %+v", loaded.HTTP)
	}

	if loaded.GRPC != nil {
		t.Errorf("expected GRPC options to be nil for an http domain, got: %+v", loaded.GRPC)
	}

	if loaded.Headers["X-Test"] != "value" {
		t.Errorf("expected shared Headers to round-trip, got: %+v", loaded.Headers)
	}
}

func TestSetAndLoadDomain_GRPC(t *testing.T) {
	tempDir := t.TempDir()
	CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	domain := &Domain{
		Name: "grpc-api",
		Base: "localhost",
		GRPC: &GRPCOptions{
			Insecure: true,
			Port:     50051,
		},
	}

	SetDomain(domain)

	loaded, err := LoadDomain("grpc-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !loaded.IsGRPC() {
		t.Errorf("expected grpc domain, IsGRPC() returned false")
	}

	if loaded.GRPC == nil || !loaded.GRPC.Insecure {
		t.Errorf("expected GRPC options to round-trip with Insecure=true, got: %+v", loaded.GRPC)
	}

	if loaded.GRPC == nil || loaded.GRPC.Port != 50051 {
		t.Errorf("expected GRPC options to round-trip with Port=50051, got: %+v", loaded.GRPC)
	}

	if loaded.HTTP != nil {
		t.Errorf("expected HTTP options to be nil for a grpc domain, got: %+v", loaded.HTTP)
	}
}

func TestDomain_HasAuth(t *testing.T) {
	tests := map[string]struct {
		domain *Domain
		want   bool
	}{
		"nil domain":          {nil, false},
		"no auth set":         {&Domain{}, false},
		"empty command":       {&Domain{Auth: &AuthOptions{}}, false},
		"command configured":  {&Domain{Auth: &AuthOptions{Command: "echo hi"}}, true},
		"command with header": {&Domain{Auth: &AuthOptions{Command: "echo hi", Header: "authorization"}}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.domain.HasAuth(); got != tc.want {
				t.Errorf("HasAuth() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSetAndLoadDomain_Auth(t *testing.T) {
	tempDir := t.TempDir()
	CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	domain := &Domain{
		Name: "auth-api",
		Base: "https://api.example.com",
		Auth: &AuthOptions{
			Command: "echo token",
			Header:  "authorization",
		},
	}

	SetDomain(domain)

	loaded, err := LoadDomain("auth-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Auth == nil {
		t.Fatalf("expected Auth options to round-trip, got nil")
	}

	if loaded.Auth.Command != "echo token" || loaded.Auth.Header != "authorization" {
		t.Errorf("Auth options did not round-trip correctly, got: %+v", loaded.Auth)
	}
}

func TestSetDomainHeader(t *testing.T) {
	tempDir := t.TempDir()
	CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	SetDomain(&Domain{Name: "a", Base: "https://a.example.com"})
	SetDomain(&Domain{Name: "b", Base: "https://b.example.com", Headers: map[string]string{"X-Existing": "kept"}})
	// SetDomain makes its argument active; re-assert "a" is active for the test.
	SetActiveName("a")

	if err := SetDomainHeader("b", "authorization", "Bearer xyz"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := LoadConfig()
	if cfg.Active != "a" {
		t.Errorf("expected active domain to remain %q, got %q", "a", cfg.Active)
	}

	loaded, err := LoadDomain("b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}

	if loaded.Headers["X-Existing"] != "kept" {
		t.Errorf("expected existing header to be preserved, got: %+v", loaded.Headers)
	}

	if loaded.Base != "https://b.example.com" {
		t.Errorf("expected Base to be untouched, got %q", loaded.Base)
	}
}

func TestSetDomainHeader_UnknownDomain(t *testing.T) {
	tempDir := t.TempDir()
	CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	if err := SetDomainHeader("does-not-exist", "authorization", "value"); err == nil {
		t.Fatal("expected an error for an unknown domain, got nil")
	}
}
