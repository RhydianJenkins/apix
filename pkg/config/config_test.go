package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultCfgPath_HonoursAPIXConfigEnvVar(t *testing.T) {
	t.Setenv("APIX_CONFIG", "/tmp/from-env.yaml")

	if got := defaultCfgPaths(); !slices.Equal(got, []string{"/tmp/from-env.yaml"}) {
		t.Errorf("defaultCfgPaths() = %q, want %q", got, "/tmp/from-env.yaml")
	}
}

func TestDefaultCfgPath_FallsBackToHomeDir(t *testing.T) {
	t.Setenv("APIX_CONFIG", "")
	t.Setenv("HOME", "/home/testuser")

	want := filepath.Join("/home/testuser", ".apix.yaml")
	if got := defaultCfgPaths(); !slices.Equal(got, []string{want}) {
		t.Errorf("defaultCfgPaths() = %q, want %q", got, want)
	}
}

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
	CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	domain := &Domain{
		Name:    "http-api",
		Base:    "https://api.example.com",
		Headers: map[string]string{"x-test": "value"},
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

	if loaded.Headers["x-test"] != "value" {
		t.Errorf("expected shared Headers to round-trip, got: %+v", loaded.Headers)
	}
}

func TestSetAndLoadDomain_GRPC(t *testing.T) {
	tempDir := t.TempDir()
	CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
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
		"nil domain":         {nil, false},
		"no auth set":        {&Domain{}, false},
		"empty command":      {&Domain{Auth: &AuthOptions{}}, false},
		"command configured": {&Domain{Auth: &AuthOptions{Command: "echo hi"}}, true},
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
	CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	domain := &Domain{
		Name: "auth-api",
		Base: "https://api.example.com",
		Auth: &AuthOptions{
			Command: "echo token",
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

	if loaded.Auth.Command != "echo token" {
		t.Errorf("Auth options did not round-trip correctly, got: %+v", loaded.Auth)
	}
}

func TestSetDomainHeader(t *testing.T) {
	tempDir := t.TempDir()
	CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	SetDomain(&Domain{Name: "a", Base: "https://a.example.com"})
	SetDomain(&Domain{Name: "b", Base: "https://b.example.com", Headers: map[string]string{"x-existing": "kept"}})
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

	if loaded.Headers["x-existing"] != "kept" {
		t.Errorf("expected existing header to be preserved, got: %+v", loaded.Headers)
	}

	if loaded.Base != "https://b.example.com" {
		t.Errorf("expected Base to be untouched, got %q", loaded.Base)
	}
}

func TestSetDomainHeader_UnknownDomain(t *testing.T) {
	tempDir := t.TempDir()
	CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	if err := SetDomainHeader("does-not-exist", "authorization", "value"); err == nil {
		t.Fatal("expected an error for an unknown domain, got nil")
	}
}

func TestDefaultCfgPaths_SplitsAPIXConfigPathList(t *testing.T) {
	t.Setenv("APIX_CONFIG", strings.Join([]string{"/a.yaml", "/b.yaml"}, string(os.PathListSeparator)))

	if got := defaultCfgPaths(); !slices.Equal(got, []string{"/a.yaml", "/b.yaml"}) {
		t.Errorf("defaultCfgPaths() = %q, want [/a.yaml /b.yaml]", got)
	}
}

// layeredFiles points CfgPaths at a tracked file and a private file, with
// the tracked one defining the domain.
func layeredFiles(t *testing.T) (tracked, private string) {
	t.Helper()

	dir := t.TempDir()
	tracked = filepath.Join(dir, ".apix.yaml")
	private = filepath.Join(dir, ".apix.private.yaml")

	trackedYAML := "active: a\ndomains:\n  a:\n    base: https://a.example.com\n    headers:\n      x-team: one\n"
	if err := os.WriteFile(tracked, []byte(trackedYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	CfgPaths = []string{tracked, private}

	return tracked, private
}

func TestLoadConfig_MultipleFiles_LaterOverridesAndDeepMerges(t *testing.T) {
	_, private := layeredFiles(t)

	privateYAML := "domains:\n  a:\n    base: https://override.example.com\n    headers:\n      authorization: secret\n  b:\n    base: https://b.example.com\n"
	if err := os.WriteFile(private, []byte(privateYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig()

	a := cfg.Domains["a"]
	if a.Base != "https://override.example.com" {
		t.Errorf("base = %q, want override", a.Base)
	}
	if a.Headers["x-team"] != "one" || a.Headers["authorization"] != "secret" {
		t.Errorf("headers = %v, want x-team and authorization merged", a.Headers)
	}
	if _, ok := cfg.Domains["b"]; !ok {
		t.Error("expected domain b from the later file")
	}
	if cfg.Active != "a" {
		t.Errorf("active = %q, want a", cfg.Active)
	}
}

func TestSetDomainHeader_MultipleFiles_WritesOnlyToLastFile(t *testing.T) {
	tracked, private := layeredFiles(t)

	before, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}

	if err := SetDomainHeader("a", "authorization", "Bearer x"); err != nil {
		t.Fatalf("SetDomainHeader() error = %v", err)
	}

	after, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("tracked file changed:\n%s", after)
	}

	privateContent, err := os.ReadFile(private)
	if err != nil {
		t.Fatalf("expected private file to be created: %v", err)
	}
	if strings.Contains(string(privateContent), "base:") {
		t.Errorf("private file should only hold the headers delta:\n%s", privateContent)
	}

	d, err := LoadDomain("a")
	if err != nil {
		t.Fatal(err)
	}
	if d.Headers["authorization"] != "Bearer x" || d.Headers["x-team"] != "one" {
		t.Errorf("merged headers = %v", d.Headers)
	}
}

func TestRemoveDomain_DefinedOnlyInEarlierFile_Errors(t *testing.T) {
	tracked, _ := layeredFiles(t)
	if err := os.WriteFile(tracked, []byte("active: other\ndomains:\n  a:\n    base: https://a.example.com\n  other:\n    base: https://o.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RemoveDomain("a"); err == nil {
		t.Fatal("expected an error removing a domain defined in another file")
	}
}

func TestLoadConfig_MissingEarlierFile_IsSkippedNotCreated(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.yaml")
	last := filepath.Join(dir, "last.yaml")
	if err := os.WriteFile(last, []byte("active: a\ndomains:\n  a:\n    base: https://a.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	CfgPaths = []string{missing, last}

	if cfg := LoadConfig(); cfg.Active != "a" {
		t.Errorf("active = %q, want a", cfg.Active)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("expected earlier missing file not to be created, stat err = %v", err)
	}
}

func TestSetDomain_MultipleFiles_WritesOnlyToLastFile(t *testing.T) {
	tracked, private := layeredFiles(t)

	before, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}

	if err := SetDomain(&Domain{Name: "b", Base: "https://b.example.com"}); err != nil {
		t.Fatalf("SetDomain() error = %v", err)
	}

	after, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("tracked file changed:\n%s", after)
	}

	privateContent, err := os.ReadFile(private)
	if err != nil {
		t.Fatalf("expected private file to be created: %v", err)
	}
	if strings.Contains(string(privateContent), "https://a.example.com") {
		t.Errorf("private file should not contain the tracked domain:\n%s", privateContent)
	}

	cfg := LoadConfig()
	if cfg.Active != "b" {
		t.Errorf("active = %q, want b", cfg.Active)
	}
	if _, ok := cfg.Domains["a"]; !ok {
		t.Error("expected domain a from the tracked file to still be present")
	}
}

func TestSetActiveName_MultipleFiles_WritesOnlyToLastFile(t *testing.T) {
	tracked, private := layeredFiles(t)

	if err := os.WriteFile(private, []byte("domains:\n  b:\n    base: https://b.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}

	if err := SetActiveName("b"); err != nil {
		t.Fatalf("SetActiveName() error = %v", err)
	}

	after, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("tracked file changed:\n%s", after)
	}

	if got := LoadConfig().Active; got != "b" {
		t.Errorf("active = %q, want b", got)
	}
}

func TestRemoveDomain_MultipleFiles_RemovesFromLastFileOnly(t *testing.T) {
	tracked, private := layeredFiles(t)

	if err := os.WriteFile(private, []byte("domains:\n  b:\n    base: https://b.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}

	if err := RemoveDomain("b"); err != nil {
		t.Fatalf("RemoveDomain() error = %v", err)
	}

	after, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("tracked file changed:\n%s", after)
	}

	cfg := LoadConfig()
	if _, ok := cfg.Domains["b"]; ok {
		t.Error("expected domain b to be removed")
	}
	if _, ok := cfg.Domains["a"]; !ok {
		t.Error("expected domain a to remain")
	}
}

func TestLoadConfig_MissingLastFile_IsCreated(t *testing.T) {
	_, private := layeredFiles(t)

	LoadConfig()

	if _, err := os.Stat(private); err != nil {
		t.Errorf("expected last file to be created: %v", err)
	}
}

func TestSetDomain_UnwritableLastFile_ReturnsError(t *testing.T) {
	CfgPaths = []string{filepath.Join(t.TempDir(), "missing-dir", ".apix.yaml")}

	if err := SetDomain(&Domain{Name: "a", Base: "https://a.example.com"}); err == nil {
		t.Fatal("expected an error when the config file cannot be created")
	}
}

func TestDefaultCfgPaths_ExpandsTildeInAPIXConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("APIX_CONFIG", strings.Join([]string{"~/a.yaml", "~/b.yaml"}, string(os.PathListSeparator)))

	want := []string{filepath.Join(home, "a.yaml"), filepath.Join(home, "b.yaml")}
	if got := defaultCfgPaths(); !slices.Equal(got, want) {
		t.Errorf("defaultCfgPaths() = %q, want %q", got, want)
	}
}
