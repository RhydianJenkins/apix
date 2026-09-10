package handlers

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/grpcclient"
)

// GRPCHandler invokes a unary gRPC method (with an empty request) on the
// given domain and returns the response as pretty-printed JSON. It merges
// the domain's default headers with the CLI-supplied ones (CLI headers take
// precedence) and sends them as gRPC metadata, before delegating dialing,
// method resolution, and invocation to pkg/grpcclient.
func GRPCHandler(domain *config.Domain, method string, headers map[string]string) ([]byte, error) {
	mergedHeaders := make(map[string]string)

	if domain != nil && domain.Headers != nil {
		maps.Copy(mergedHeaders, domain.Headers)
	}

	maps.Copy(mergedHeaders, headers)

	conn, err := grpcclient.Dial(domain)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	methodInfo, err := grpcclient.ResolveMethod(ctx, conn, method)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve method %q: %w", method, err)
	}

	body, err := grpcclient.Invoke(ctx, conn, methodInfo, mergedHeaders)
	if err != nil {
		return nil, err
	}

	return body, nil
}
