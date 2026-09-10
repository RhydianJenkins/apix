package handlers

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/grpcclient"
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
		return nil, err
	}

	return body, nil
}
