package grpcclient_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rhydianjenkins/apix/pkg/grpcclient"
)

func TestInvoke(t *testing.T) {
	srv, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := grpcclient.ResolveMethod(ctx, conn, "GetReply", nil)
	if err != nil {
		t.Fatalf("failed to resolve method: %v", err)
	}

	body, err := grpcclient.Invoke(ctx, conn, info, nil, map[string]string{"x-test": "abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(body), "hello") {
		t.Errorf("expected response to contain %q, got: %s", "hello", string(body))
	}

	md := srv.LastMetadata()
	values := md.Get("x-test")
	if len(values) == 0 || values[0] != "abc" {
		t.Errorf("expected metadata x-test=abc to reach the server, got: %v", values)
	}
}

func TestInvoke_WithRequestBody(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := grpcclient.ResolveMethod(ctx, conn, "GetReply", nil)
	if err != nil {
		t.Fatalf("failed to resolve method: %v", err)
	}

	reqBody := []byte(`{"name": "world"}`)

	body, err := grpcclient.Invoke(ctx, conn, info, &reqBody, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(string(body), "hello, world") {
		t.Errorf("expected response to contain %q, got: %s", "hello, world", string(body))
	}
}

func TestInvoke_WithBlankRequestBody(t *testing.T) {
	srv, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := grpcclient.ResolveMethod(ctx, conn, "GetReply", nil)
	if err != nil {
		t.Fatalf("failed to resolve method: %v", err)
	}

	reqBody := []byte("   \n")

	if _, err := grpcclient.Invoke(ctx, conn, info, &reqBody, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := srv.LastName(); got != "" {
		t.Errorf("expected a blank request body to be treated as empty, got name=%q", got)
	}
}

func TestInvoke_WithInvalidRequestBody(t *testing.T) {
	_, conn := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := grpcclient.ResolveMethod(ctx, conn, "GetReply", nil)
	if err != nil {
		t.Fatalf("failed to resolve method: %v", err)
	}

	reqBody := []byte(`{not valid json`)

	_, err = grpcclient.Invoke(ctx, conn, info, &reqBody, nil)
	if err == nil {
		t.Fatal("expected an error for an invalid JSON request body, got nil")
	}
}
