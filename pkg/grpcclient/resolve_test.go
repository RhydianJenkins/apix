package grpcclient_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rhydianjenkins/apix/internal/grpctest"
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/grpcclient"
	"google.golang.org/grpc"
)

func startTestServer(t *testing.T) (*grpctest.Server, *grpc.ClientConn) {
	t.Helper()

	srv, err := grpctest.NewServer()
	if err != nil {
		t.Fatalf("failed to build test server: %v", err)
	}
	t.Cleanup(srv.Close)

	go srv.Serve()

	domain := &config.Domain{
		Base:     srv.Host(),
		Name:     "testgrpc",
		Protocol: config.ProtocolGRPC,
		GRPC:     &config.GRPCOptions{Insecure: true, Port: srv.Port()},
	}

	conn, err := grpcclient.Dial(domain)
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return srv, conn
}

func TestResolveMethod_BareName(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := grpcclient.ResolveMethod(ctx, conn, grpctest.MethodName, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFullName := "/" + grpctest.ServiceName + "/" + grpctest.MethodName
	if info.FullName != wantFullName {
		t.Errorf("expected FullName %q, got %q", wantFullName, info.FullName)
	}

	if info.IsClientStreaming || info.IsServerStreaming {
		t.Errorf("expected a unary method, got streaming flags %v/%v", info.IsClientStreaming, info.IsServerStreaming)
	}
}

func TestResolveMethod_Qualified(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	qualified := grpctest.ServiceName + "/" + grpctest.MethodName
	info, err := grpcclient.ResolveMethod(ctx, conn, qualified, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.FullName != "/"+qualified {
		t.Errorf("expected FullName %q, got %q", "/"+qualified, info.FullName)
	}
}

func TestResolveMethod_NotFound(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := grpcclient.ResolveMethod(ctx, conn, "DoesNotExist", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown method, got nil")
	}

	if !strings.Contains(err.Error(), "no method named") {
		t.Errorf("expected a 'no method named' error, got: %v", err)
	}
}

func TestResolveMethod_ExcludesReflectionService(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := grpcclient.ResolveMethod(ctx, conn, "ServerReflectionInfo", nil)
	if err == nil {
		t.Fatal("expected the reflection service's own method to be excluded from resolution")
	}
}

func TestListMethods(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	methods, err := grpcclient.ListMethods(ctx, conn, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := grpctest.ServiceName + "/" + grpctest.MethodName
	if len(methods) != 1 || methods[0] != want {
		t.Errorf("expected methods %v, got %v", []string{want}, methods)
	}
}

// Reflection is a separate RPC from the eventual method call - servers that
// require auth for regular RPCs typically require it for reflection too, so
// metadata passed to ResolveMethod must reach the reflection stream itself.
func TestResolveMethod_SendsMetadata(t *testing.T) {
	srv, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := grpcclient.ResolveMethod(ctx, conn, grpctest.MethodName, map[string]string{"authorization": "Bearer test-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := srv.LastMetadata()
	values := md.Get("authorization")
	if len(values) == 0 || values[0] != "Bearer test-token" {
		t.Errorf("expected metadata authorization=Bearer test-token to reach the reflection call, got: %v", values)
	}
}
