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

func newStatusTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("verbose", "v", false, "")
	return cmd
}

func TestStatusHandler_PrintsOnlyActiveDomainName_ByDefault(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{Name: "my-api", Base: "https://api.example.com"})
	config.SetDomain(&config.Domain{Name: "other-api", Base: "https://other.example.com"})
	config.SetActiveName("my-api")

	output := captureStdout(t, func() {
		if err := StatusHandler(newStatusTestCmd(), []string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.TrimSpace(output) != "my-api" {
		t.Errorf("expected output to be just the active domain name, got:\n%s", output)
	}
}

func TestStatusHandler_Verbose_PrintsActiveDomainDetails(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	config.SetDomain(&config.Domain{
		Name:    "my-api",
		Base:    "https://api.example.com",
		Headers: map[string]string{"x-test": "value"},
		HTTP:    &config.HTTPOptions{User: "foo", Pass: "bar"},
	})
	config.SetDomain(&config.Domain{Name: "other-api", Base: "https://other.example.com"})
	config.SetActiveName("my-api")

	cmd := newStatusTestCmd()
	cmd.Flags().Set("verbose", "true")

	output := captureStdout(t, func() {
		if err := StatusHandler(cmd, []string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{"my-api", "https://api.example.com", "http", "foo", "bar", "x-test: value"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, output)
		}
	}

	if strings.Contains(output, "other-api") {
		t.Errorf("expected output to not mention the inactive domain, got:\n%s", output)
	}
}

func TestStatusHandler_NoActiveDomain(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPaths = []string{filepath.Join(tempDir, ".apix.yaml")}
	viper.Reset()

	if err := StatusHandler(newStatusTestCmd(), []string{}); err == nil {
		t.Fatal("expected an error when no active domain is set, got nil")
	}
}

func TestStatusHandler_ActiveDomainRemovedFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, ".apix.yaml")
	config.CfgPaths = []string{cfgPath}
	viper.Reset()

	// simulate an active domain that is no longer defined in config, without
	// going through RemoveDomain (which refuses to remove the active domain)
	yamlContent := "active: my-api\ndomains: {}\n"
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if err := StatusHandler(newStatusTestCmd(), []string{}); err == nil {
		t.Fatal("expected an error for a dangling active domain, got nil")
	}
}
