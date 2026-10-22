package handlers

import (
	"context"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/grpcclient"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCShowHandler resolves a gRPC method via server reflection and returns a
// human-readable summary of its request/response message shapes, without
// invoking it.
func GRPCShowHandler(domain *config.Domain, method string, headers map[string]string) (string, error) {
	mergedHeaders := make(map[string]string)

	if domain != nil && domain.Headers != nil {
		maps.Copy(mergedHeaders, domain.Headers)
	}

	maps.Copy(mergedHeaders, headers)

	conn, err := grpcclient.Dial(domain)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	methodInfo, err := grpcclient.ResolveMethod(ctx, conn, method, mergedHeaders)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated && domain.HasAuth() {
			fmt.Fprintf(os.Stderr, "hint: run `apix auth %s` to refresh credentials\n", domain.Name)
		}
		return "", fmt.Errorf("failed to resolve method %q: %w", method, err)
	}

	return grpcclient.FormatMethodSummary(methodInfo), nil
}
