package handlers

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAuthTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("all", false, "")
	return cmd
}

func TestRunAuth_CapturesHeader(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	domain := &config.Domain{
		Name: "auth-api",
		Base: "https://api.example.com",
		Auth: &config.AuthOptions{
			Command: "printf token-value",
			Header:  "authorization",
		},
	}
	config.SetDomain(domain)

	if err := runAuth(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := config.LoadDomain("auth-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "token-value" {
		t.Errorf("expected authorization header to be set to %q, got: %+v", "token-value", loaded.Headers)
	}
}

func TestRunAuth_NoHeader_SideEffectOnly(t *testing.T) {
	domain := &config.Domain{
		Name: "side-effect-api",
		Auth: &config.AuthOptions{Command: "true"},
	}

	if err := runAuth(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunAuth_CommandFails(t *testing.T) {
	domain := &config.Domain{
		Name: "failing-api",
		Auth: &config.AuthOptions{Command: "exit 1"},
	}

	err := runAuth(domain)
	if err == nil {
		t.Fatal("expected an error for a failing auth command, got nil")
	}

	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("expected error to mention failure, got: %v", err)
	}
}

func TestRunAuth_EmptyOutputWithHeaderSet(t *testing.T) {
	domain := &config.Domain{
		Name: "empty-output-api",
		Auth: &config.AuthOptions{Command: "true", Header: "authorization"},
	}

	err := runAuth(domain)
	if err == nil {
		t.Fatal("expected an error for empty output with a header configured, got nil")
	}

	if !strings.Contains(err.Error(), "produced no output") {
		t.Errorf("expected error to mention empty output, got: %v", err)
	}
}

func TestRunAuth_NoAuthConfigured(t *testing.T) {
	domain := &config.Domain{Name: "no-auth-api"}

	err := runAuth(domain)
	if err == nil {
		t.Fatal("expected an error for a domain with no auth configured, got nil")
	}

	if !strings.Contains(err.Error(), "no auth command configured") {
		t.Errorf("expected error to mention missing auth command, got: %v", err)
	}
}

func TestAuthHandler_NamedDomain_DoesNotChangeActive(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{
		Name: "b",
		Base: "https://b.example.com",
		Auth: &config.AuthOptions{Command: "printf b-token", Header: "authorization"},
	})
	config.SetActiveName("a")

	cmd := newAuthTestCmd()

	if err := AuthHandler(cmd, []string{"b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := config.LoadConfig().Active; got != "a" {
		t.Errorf("expected active domain to remain %q, got %q", "a", got)
	}

	loaded, err := config.LoadDomain("b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "b-token" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
	}
}

func TestAuthHandler_All(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "no-auth", Base: "https://no-auth.example.com"})
	config.SetDomain(&config.Domain{
		Name: "good",
		Base: "https://good.example.com",
		Auth: &config.AuthOptions{Command: "printf good-token", Header: "authorization"},
	})
	config.SetDomain(&config.Domain{
		Name: "bad",
		Base: "https://bad.example.com",
		Auth: &config.AuthOptions{Command: "exit 1"},
	})

	cmd := newAuthTestCmd()
	cmd.Flags().Set("all", "true")

	err := AuthHandler(cmd, nil)
	if err == nil {
		t.Fatal("expected an error since one domain's auth command fails, got nil")
	}

	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("expected error to mention the failing domain %q, got: %v", "bad", err)
	}

	loaded, loadErr := config.LoadDomain("good")
	if loadErr != nil {
		t.Fatalf("unexpected error: %v", loadErr)
	}

	if loaded.Headers["authorization"] != "good-token" {
		t.Errorf("expected the succeeding domain to still have its header set, got: %+v", loaded.Headers)
	}
}

func TestAuthHandler_AllWithDomainArg_Errors(t *testing.T) {
	cmd := newAuthTestCmd()
	cmd.Flags().Set("all", "true")

	err := AuthHandler(cmd, []string{"some-domain"})
	if err == nil {
		t.Fatal("expected an error when combining --all with a domain name, got nil")
	}
}
