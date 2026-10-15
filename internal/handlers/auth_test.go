package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newAuthTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("all", false, "")
	return cmd
}

func TestRunAuth_SetsDomainEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "domain.txt")

	domain := &config.Domain{
		Name: "auth-api",
		Base: "https://api.example.com",
		Auth: &config.AuthOptions{
			Command: "printf '%s' \"$APIX_DOMAIN\" > " + outFile,
		},
	}

	if err := runAuth(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("unexpected error reading command output: %v", err)
	}

	if string(got) != "auth-api" {
		t.Errorf("expected APIX_DOMAIN to be %q, got %q", "auth-api", string(got))
	}
}

func TestRunAuth_SetsConfigEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, "custom.yaml")}
	viper.Reset()

	outFile := filepath.Join(tempDir, "config-path.txt")

	domain := &config.Domain{
		Name: "auth-api",
		Auth: &config.AuthOptions{
			Command: "printf '%s' \"$APIX_CONFIG\" > " + outFile,
		},
	}

	if err := runAuth(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("unexpected error reading command output: %v", err)
	}

	if want := strings.Join(config.CfgPaths, string(os.PathListSeparator)); string(got) != want {
		t.Errorf("expected APIX_CONFIG to be %q, got %q", want, string(got))
	}
}

func TestRunAuth_RunsSideEffectCommand(t *testing.T) {
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
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	outFile := filepath.Join(tempDir, "domain.txt")

	config.SetDomain(&config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{
		Name: "b",
		Base: "https://b.example.com",
		Auth: &config.AuthOptions{Command: "printf '%s' \"$APIX_DOMAIN\" > " + outFile},
	})
	config.SetActiveName("a")

	cmd := newAuthTestCmd()

	if err := AuthHandler(cmd, []string{"b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := config.LoadConfig().Active; got != "a" {
		t.Errorf("expected active domain to remain %q, got %q", "a", got)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("unexpected error reading command output: %v", err)
	}

	if string(got) != "b" {
		t.Errorf("expected APIX_DOMAIN to be %q, got %q", "b", string(got))
	}
}

func TestAuthHandler_All(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	goodMarker := filepath.Join(tempDir, "good.txt")

	config.SetDomain(&config.Domain{Name: "no-auth", Base: "https://no-auth.example.com"})
	config.SetDomain(&config.Domain{
		Name: "good",
		Base: "https://good.example.com",
		Auth: &config.AuthOptions{Command: "touch " + goodMarker},
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

	if _, err := os.Stat(goodMarker); err != nil {
		t.Errorf("expected the succeeding domain's command to have run, got: %v", err)
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

func TestRunAuth_SetsConfigEnvVar_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{
		filepath.Join(tempDir, "tracked.yaml"),
		filepath.Join(tempDir, "private.yaml"),
	}

	outFile := filepath.Join(tempDir, "config-path.txt")

	domain := &config.Domain{
		Name: "auth-api",
		Auth: &config.AuthOptions{
			Command: "printf '%s' \"$APIX_CONFIG\" > " + outFile,
		},
	}

	if err := runAuth(domain); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("unexpected error reading command output: %v", err)
	}

	want := config.CfgPaths[0] + string(os.PathListSeparator) + config.CfgPaths[1]
	if string(got) != want {
		t.Errorf("expected APIX_CONFIG to be %q, got %q", want, string(got))
	}
}
