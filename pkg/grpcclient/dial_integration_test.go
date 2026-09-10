package grpcclient_test

import (
	"testing"

	"github.com/rhydianjenkins/apix/internal/grpctest"
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/grpcclient"
)

func TestDial_WithSeparatePort(t *testing.T) {
	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	t.Cleanup(srv.Close)

	go srv.Serve()

	domain := &config.Domain{
		Base: srv.Host(),
		GRPC: &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	}

	conn, err := grpcclient.Dial(domain)
	if err != nil {
		t.Fatalf("failed to dial with a separate port: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
}

func TestDial_PortInBase(t *testing.T) {
	domain := &config.Domain{
		Base: "localhost:1234",
		GRPC: &config.GRPCOptions{Insecure: true, Port: 50051},
	}

	if _, err := grpcclient.Dial(domain); err == nil {
		t.Fatal("expected an error when base includes a port")
	}
}

func TestDial_MissingPort(t *testing.T) {
	domain := &config.Domain{
		Base: "localhost",
		GRPC: &config.GRPCOptions{Insecure: true},
	}

	if _, err := grpcclient.Dial(domain); err == nil {
		t.Fatal("expected an error when no port is configured")
	}
}
