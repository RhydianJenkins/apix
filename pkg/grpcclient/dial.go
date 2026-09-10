package grpcclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

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

	target, err := grpcTarget(domain)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial %q: %w", target, err)
	}

	return conn, nil
}

// grpcTarget returns the dial target for domain: a host-only domain.Base
// combined with domain.GRPC.Port (e.g. "localhost" + 50051 ->
// "localhost:50051"). base must never include a port itself - the port is
// always specified separately via --port / grpc.port.
func grpcTarget(domain *config.Domain) (string, error) {
	if strings.Contains(domain.Base, ":") {
		return "", fmt.Errorf("base %q must not include a port; set it separately with --port or grpc.port instead", domain.Base)
	}

	if domain.GRPC == nil || domain.GRPC.Port == 0 {
		return "", fmt.Errorf("domain %q has no grpc port configured; set it with --port or grpc.port in %s", domain.Name, config.CfgPath)
	}

	return net.JoinHostPort(domain.Base, strconv.Itoa(domain.GRPC.Port)), nil
}
