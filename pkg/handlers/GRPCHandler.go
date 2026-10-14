package handlers

import (
	"context"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/rhydianjenkins/apix/v3/pkg/config"
	"github.com/rhydianjenkins/apix/v3/pkg/grpcclient"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCHandler(domain *config.Domain, method string, reqBody *[]byte, headers map[string]string) ([]byte, error) {
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

	methodInfo, err := grpcclient.ResolveMethod(ctx, conn, method, mergedHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve method %q: %w", method, err)
	}

	body, err := grpcclient.Invoke(ctx, conn, methodInfo, reqBody, mergedHeaders)
	if err != nil {
		if status.Code(err) == codes.Unauthenticated && domain.HasAuth() {
			fmt.Fprintf(os.Stderr, "hint: run `apix auth %s` to refresh credentials\n", domain.Name)
		}
		return nil, err
	}

	return body, nil
}
