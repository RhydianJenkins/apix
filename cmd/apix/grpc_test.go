package main

import (
	"testing"

	"github.com/rhydianjenkins/apix/internal/grpctest"
	"github.com/rhydianjenkins/apix/pkg/config"
)

func TestGRPCCommand_UnknownMethodReturnsError(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: srv.Host(),
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"grpc", "NoSuchMethod"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error for an unknown grpc method, got nil")
	}
}

func TestGRPCCommand_SuccessReturnsNilError(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	setActiveDomain(t, &config.Domain{
		Name: "testgrpc",
		Base: srv.Host(),
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"grpc", grpctest.MethodName})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected Execute() to return nil for a successful grpc call, got: %v", err)
	}
}

func TestGRPCCommand_WrongProtocolReturnsError(t *testing.T) {
	setActiveDomain(t, &config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
	})

	cmd := newRootCmd("test")
	cmd.SetArgs([]string{"grpc", "SomeMethod"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute() to return an error when the active domain is not grpc, got nil")
	}
}
