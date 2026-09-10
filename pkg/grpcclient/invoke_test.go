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

	body, err := grpcclient.Invoke(ctx, conn, info, map[string]string{"x-test": "abc"})
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
