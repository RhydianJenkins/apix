package main

import (
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
)

func TestNewCommand_GraphQLFlag(t *testing.T) {
	setActiveDomain(t, &config.Domain{Name: "placeholder", Base: "https://placeholder.example.com"})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{
		"new", "testgql", "https://api.example.com/graphql",
		"--graphql",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil, got: %v", err)
	}

	loaded, err := config.LoadDomain("testgql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loaded.HTTP == nil || !loaded.HTTP.GraphQL {
		t.Errorf("expected HTTP.GraphQL to be true, got: %+v", loaded.HTTP)
	}

	if !loaded.IsGraphQL() {
		t.Errorf("expected IsGraphQL() to be true for the saved domain")
	}
}
