package grpcclient

import (
	"crypto/tls"
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

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
