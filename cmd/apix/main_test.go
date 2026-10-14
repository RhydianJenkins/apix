package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/pkg/config"
	"github.com/spf13/viper"
)

func TestRootCmd_ConfigFlag_OverridesCfgPath(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "custom.yaml")

	config.CfgPath = filepath.Join(tempDir, "unused.yaml") // sanity: should be overridden, not used
	viper.Reset()

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"-c", customPath, "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	if config.CfgPath != customPath {
		t.Fatalf("expected CfgPath to be %q, got %q", customPath, config.CfgPath)
	}

	if _, err := os.Stat(customPath); err != nil {
		t.Fatalf("expected config file to be created at %q: %v", customPath, err)
	}
}

func TestRootCmd_ConfigFlag_ExpandsTilde(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	viper.Reset()

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"-c", "~/custom.yaml", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	want := filepath.Join(fakeHome, "custom.yaml")
	if config.CfgPath != want {
		t.Fatalf("expected CfgPath to be %q, got %q", want, config.CfgPath)
	}
}

// TestRootCmd_ConfigFlag_AppliesDuringCompletion is a regression test: Cobra's
// shell completion (`__complete`) parses flags but never calls
// PersistentPreRun, so a naive implementation of --config that only applies
// the override there would silently keep using the default CfgPath during
// tab completion. This reproduces `apix -c ./custom.yaml get <TAB>`.
func TestRootCmd_ConfigFlag_AppliesDuringCompletion(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	tempDir := t.TempDir()
	specPath := writeTestOAS(t)
	customConfigPath := filepath.Join(tempDir, "custom.yaml")

	config.CfgPath = customConfigPath
	viper.Reset()
	config.SetDomain(&config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{OpenAPISpecPath: specPath},
	})

	// Point the "default" CfgPath somewhere with no valid OAS spec, so this
	// test fails loudly if completion silently falls back to it instead of
	// honouring -c.
	config.CfgPath = filepath.Join(tempDir, "default.yaml")
	viper.Reset()

	cmd := newRootCmd("test")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"__complete", "-c", customConfigPath, "get", ""})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	got := out.String()
	for _, want := range []string{"/users", "/users/{id}"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected completion output to contain %q, got:\n%s", want, got)
		}
	}
}

func TestRootCmd_NoConfigFlag_LeavesCfgPathUnchanged(t *testing.T) {
	t.Cleanup(func() { viper.Reset() })

	tempDir := t.TempDir()
	existing := filepath.Join(tempDir, ".apix.yaml")
	config.CfgPath = existing
	viper.Reset()

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	if config.CfgPath != existing {
		t.Fatalf("expected CfgPath to remain %q, got %q", existing, config.CfgPath)
	}
}
