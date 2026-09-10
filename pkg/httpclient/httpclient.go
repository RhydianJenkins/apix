// Package httpclient contains the low-level HTTP transport logic used to
// send requests to an http-protocol domain. It knows nothing about cobra,
// config loading, or response formatting beyond building and executing a
// single *http.Request - that orchestration lives in pkg/handlers.
package httpclient

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
)

// Do builds and sends a single HTTP request to the given domain/path,
// applying basic auth (if configured) and the given headers on top of the
// package's sensible defaults (Content-Type/Accept/User-Agent).
func Do(
	method string,
	domain *config.Domain,
	path string,
	reqBody *[]byte,
	headers map[string]string,
) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var req *http.Request
	var err error

	url := domain.Base + path

	if reqBody != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(*reqBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if domain != nil && domain.HTTP != nil && domain.HTTP.User != "" && domain.HTTP.Pass != "" {
		req.SetBasicAuth(domain.HTTP.User, domain.HTTP.Pass)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	setDefaultHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func setDefaultHeaders(req *http.Request) {
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	}
}
