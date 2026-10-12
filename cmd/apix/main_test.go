package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
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

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to resolve home dir: %v", err)
	}

	tempDir, err := os.MkdirTemp(home, "apix-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir under home: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	rel, err := filepath.Rel(home, filepath.Join(tempDir, "custom.yaml"))
	if err != nil {
		t.Fatalf("failed to compute relative path: %v", err)
	}

	viper.Reset()

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"-c", "~/" + rel, "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	want := filepath.Join(tempDir, "custom.yaml")
	if config.CfgPath != want {
		t.Fatalf("expected CfgPath to be %q, got %q", want, config.CfgPath)
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
