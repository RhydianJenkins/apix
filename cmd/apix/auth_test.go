package main

import (
	"testing"

	"github.com/rhydianjenkins/apix/v2/pkg/config"
)

func TestAuthCommand_ActiveDomain(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testauth",
		Base: "https://api.example.com",
		Auth: &config.AuthOptions{Command: "printf cli-token", Header: "authorization"},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"auth"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	loaded, err := config.LoadDomain("testauth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["authorization"] != "cli-token" {
		t.Errorf("expected authorization header to be set, got: %+v", loaded.Headers)
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
		"--auth-header", "authorization",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	loaded, err := config.LoadDomain("testauth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Auth == nil || loaded.Auth.Command != "printf token" || loaded.Auth.Header != "authorization" {
		t.Errorf("expected Auth options to be set from flags, got: %+v", loaded.Auth)
	}
}
