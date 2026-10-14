package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/v3/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSetHeaderTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("domain", "", "")
	return cmd
}

func TestSetHeaderHandler_DomainFlag(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{Name: "b", Base: "https://b.example.com"})
	config.SetActiveName("a")

	cmd := newSetHeaderTestCmd()
	cmd.Flags().Set("domain", "b")

	if err := SetHeaderHandler(cmd, []string{"authorization", "Bearer xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := config.LoadConfig().Active; got != "a" {
		t.Errorf("expected active domain to remain %q, got %q", "a", got)
	}

	loaded, err := config.LoadDomain("b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}
}

func TestSetHeaderHandler_EnvVar(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{Name: "b", Base: "https://b.example.com"})
	config.SetActiveName("a")

	t.Setenv("APIX_DOMAIN", "b")

	cmd := newSetHeaderTestCmd()

	if err := SetHeaderHandler(cmd, []string{"authorization", "Bearer xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := config.LoadDomain("b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}
}

func TestSetHeaderHandler_DomainFlag_OverridesEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{Name: "b", Base: "https://b.example.com"})

	t.Setenv("APIX_DOMAIN", "b")

	cmd := newSetHeaderTestCmd()
	cmd.Flags().Set("domain", "a")

	if err := SetHeaderHandler(cmd, []string{"authorization", "Bearer xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := config.LoadDomain("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set on %q, got: %+v", "a", loaded.Headers)
	}
}

func TestSetHeaderHandler_FallsBackToActiveDomain(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "active-api", Base: "https://active.example.com"})

	cmd := newSetHeaderTestCmd()

	if err := SetHeaderHandler(cmd, []string{"authorization", "Bearer xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := config.LoadDomain("active-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}
}

func TestSetHeaderHandler_UnknownDomain(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	cmd := newSetHeaderTestCmd()
	cmd.Flags().Set("domain", "does-not-exist")

	if err := SetHeaderHandler(cmd, []string{"authorization", "value"}); err == nil {
		t.Fatal("expected an error for an unknown domain, got nil")
	}
}

func TestSetHeaderHandler_MultipleFiles_WritesToLastFile(t *testing.T) {
	tempDir := t.TempDir()
	tracked := filepath.Join(tempDir, ".apix.yaml")
	private := filepath.Join(tempDir, ".apix.private.yaml")

	config.CfgPaths = []string{tracked}
	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})

	before, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}

	config.CfgPaths = []string{tracked, private}

	if err := SetHeaderHandler(newSetHeaderTestCmd(), []string{"authorization", "Bearer xyz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after, err := os.ReadFile(tracked)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("tracked file changed:\n%s", after)
	}

	loaded, err := config.LoadDomain("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Headers["authorization"] != "Bearer xyz" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}
}
