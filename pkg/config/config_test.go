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
		"nil domain":               {nil, false},
		"empty protocol (default)": {&Domain{}, false},
		"explicit http":            {&Domain{Protocol: ProtocolHTTP}, false},
		"explicit grpc":            {&Domain{Protocol: ProtocolGRPC}, true},
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
		Name:     "http-api",
		Base:     "https://api.example.com",
		Protocol: ProtocolHTTP,
		Headers:  map[string]string{"X-Test": "value"},
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
		Name:     "grpc-api",
		Base:     "localhost",
		Protocol: ProtocolGRPC,
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
