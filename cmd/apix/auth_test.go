package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

func TestAuthCommand_ActiveDomain(t *testing.T) {
	tempDir := t.TempDir()
	outFile := filepath.Join(tempDir, "domain.txt")

	setActiveDomain(t, &config.Domain{
		Name: "testauth",
		Base: "https://api.example.com",
		Auth: &config.AuthOptions{Command: "printf '%s' \"$APIX_DOMAIN\" > " + outFile},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"auth"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("unexpected error reading command output: %v", err)
	}

	if string(got) != "testauth" {
		t.Errorf("expected APIX_DOMAIN to be %q, got %q", "testauth", string(got))
	}
}

func TestAuthCommand_NoAuthConfiguredReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testauth",
		Base: "https://api.example.com",
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"auth"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when no auth command is configured, got nil")
	}
}

func TestNewCommand_AuthFlags(t *testing.T) {
	setActiveDomain(t, &config.Domain{Name: "placeholder", Base: "https://placeholder.example.com"})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{
		"new", "testauth", "https://api.example.com",
		"--auth-command", "printf token",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	loaded, err := config.LoadDomain("testauth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Auth == nil || loaded.Auth.Command != "printf token" {
		t.Errorf("expected Auth options to be set from flags, got: %+v", loaded.Auth)
	}
}
