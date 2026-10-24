package main

import (
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

func TestSetHeaderCommand_DomainFlag(t *testing.T) {
	setActiveDomain(t, &config.Domain{Name: "a", Base: "https://a.example.com"})
	config.SetDomain(&config.Domain{Name: "b", Base: "https://b.example.com"})
	config.SetActiveName("a")

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"set-header", "authorization", "Bearer xyz", "--domain", "b"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.Active; got != "a" {
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

func TestSetHeaderCommand_DefaultsToActiveDomain(t *testing.T) {
	setActiveDomain(t, &config.Domain{Name: "testheader", Base: "https://api.example.com"})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"set-header", "x-my-header", "value"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	loaded, err := config.LoadDomain("testheader")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.Headers["x-my-header"] != "value" {
		t.Errorf("expected header to be set, got: %+v", loaded.Headers)
	}
}
