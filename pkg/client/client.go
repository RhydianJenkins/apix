package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
)

func MakeRequest(
	method string,
	domain *config.Domain,
	path string,
	body []byte,
	headers map[string]string,
) (*http.Response, []byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := domain.Base + path

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if domain.User != "" && domain.Pass != "" {
		req.SetBasicAuth(domain.User, domain.Pass)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, responseBody, nil
}
