package grpcclient

import (
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
)

func TestGRPCTarget(t *testing.T) {
	tests := map[string]struct {
		domain  *config.Domain
		want    string
		wantErr bool
	}{
		"host-only base with explicit port": {
			domain: &config.Domain{Base: "localhost", GRPC: &config.GRPCOptions{Port: 50051}},
			want:   "localhost:50051",
		},
		"host-only remote base with explicit port": {
			domain: &config.Domain{Base: "grpc.example.com", GRPC: &config.GRPCOptions{Port: 443}},
			want:   "grpc.example.com:443",
		},
		"no grpc options at all": {
			domain:  &config.Domain{Base: "localhost"},
			wantErr: true,
		},
		"grpc options without a port": {
			domain:  &config.Domain{Base: "localhost", GRPC: &config.GRPCOptions{Insecure: true}},
			wantErr: true,
		},
		"base includes a port": {
			domain:  &config.Domain{Base: "localhost:50051"},
			wantErr: true,
		},
		"base includes a port even when an explicit port is also set": {
			domain:  &config.Domain{Base: "localhost:1234", GRPC: &config.GRPCOptions{Port: 50051}},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := grpcTarget(tc.domain)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got target %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("grpcTarget() = %q, want %q", got, tc.want)
			}
		})
	}
}
