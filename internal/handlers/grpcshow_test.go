package handlers

import (
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/grpctest"
)

func TestGRPCShowHandler(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	domain := &config.Domain{
		Base: srv.Host(),
		Name: "testgrpc",
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	}

	summary, err := GRPCShowHandler(domain, grpctest.MethodName, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantContains := []string{
		grpctest.ServiceName + "/" + grpctest.MethodName,
		"Request: testpkg.GetReplyRequest",
		"name string",
		"Response: testpkg.Reply",
		"message string",
	}

	for _, want := range wantContains {
		if !strings.Contains(summary, want) {
			t.Errorf("expected summary to contain %q, got:\n%s", want, summary)
		}
	}
}

func TestGRPCShowHandler_UnknownMethod(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	defer srv.Close()

	go srv.Serve()

	domain := &config.Domain{
		Base: srv.Host(),
		Name: "testgrpc",
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	}

	_, err = GRPCShowHandler(domain, "NoSuchMethod", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown method, got nil")
	}
}
