package httpclient

import (
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
)

func TestTarget(t *testing.T) {
	tests := map[string]struct {
		domain  *config.Domain
		want    string
		wantErr bool
	}{
		"no http options": {
			domain: &config.Domain{Base: "https://api.example.com"},
			want:   "https://api.example.com",
		},
		"http options without a port": {
			domain: &config.Domain{Base: "https://api.example.com", HTTP: &config.HTTPOptions{User: "foo"}},
			want:   "https://api.example.com",
		},
		"host-only base with explicit port": {
			domain: &config.Domain{Base: "https://api.example.com", HTTP: &config.HTTPOptions{Port: 8443}},
			want:   "https://api.example.com:8443",
		},
		"host with path and explicit port": {
			domain: &config.Domain{Base: "https://api.example.com/v1", HTTP: &config.HTTPOptions{Port: 8443}},
			want:   "https://api.example.com:8443/v1",
		},
		"plain http scheme with explicit port": {
			domain: &config.Domain{Base: "http://localhost", HTTP: &config.HTTPOptions{Port: 8080}},
			want:   "http://localhost:8080",
		},
		"base already includes a port and an explicit port is set": {
			domain:  &config.Domain{Base: "https://api.example.com:443", HTTP: &config.HTTPOptions{Port: 8443}},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Target(tc.domain)

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
				t.Errorf("Target() = %q, want %q", got, tc.want)
			}
		})
	}
}
