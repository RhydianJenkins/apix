// Package grpcclient contains the low-level gRPC transport logic used to
// call a method on a grpc-protocol domain: dialing, resolving a method via
// server reflection, and invoking it. It knows nothing about cobra, config
// loading beyond *config.Domain, or response formatting for the CLI - that
// orchestration lives in pkg/handlers.
package grpcclient

import (
	"crypto/tls"
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial opens a gRPC client connection to the domain's Base address
// (host:port), using TLS with the system cert pool by default, or plaintext
// if the domain opts in via GRPC.Insecure.
func Dial(domain *config.Domain) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials

	if domain != nil && domain.GRPC != nil && domain.GRPC.Insecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{})
	}

	conn, err := grpc.NewClient(domain.Base, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial %q: %w", domain.Base, err)
	}

	return conn, nil
}
