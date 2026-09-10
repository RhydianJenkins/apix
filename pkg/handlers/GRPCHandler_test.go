package handlers

import (
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/internal/grpctest"
	"github.com/rhydianjenkins/apix/pkg/config"
)

func TestGRPCHandler(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	domain := &config.Domain{
		Base:     srv.Host(),
		Name:     "testgrpc",
		Protocol: config.ProtocolGRPC,
		GRPC:     &config.GRPCOptions{Insecure: true, Port: srv.Port()},
		Headers: map[string]string{
			"x-config": "config-value",
		},
	}

	body, err := GRPCHandler(domain, grpctest.MethodName, map[string]string{"x-cli": "cli-value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(body), "hello") {
		t.Errorf("expected response to contain %q, got: %s", "hello", string(body))
	}

	md := srv.LastMetadata()

	if got := md.Get("x-config"); len(got) == 0 || got[0] != "config-value" {
		t.Errorf("expected domain header x-config to be sent as metadata, got: %v", got)
	}

	if got := md.Get("x-cli"); len(got) == 0 || got[0] != "cli-value" {
		t.Errorf("expected CLI header x-cli to be sent as metadata, got: %v", got)
	}
}

func TestGRPCHandler_UnknownMethod(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	domain := &config.Domain{
		Base:     srv.Host(),
		Name:     "testgrpc",
		Protocol: config.ProtocolGRPC,
		GRPC:     &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	}

	_, err = GRPCHandler(domain, "NoSuchMethod", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown method, got nil")
	}
}
