package httpclient

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
)

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

	base, err := Target(domain)
	if err != nil {
		return nil, err
	}

	reqURL := base + path

	if reqBody != nil {
		req, err = http.NewRequest(method, reqURL, bytes.NewBuffer(*reqBody))
	} else {
		req, err = http.NewRequest(method, reqURL, nil)
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

// base must never include a port itself when a port is configured
// separately - the port is always specified via --port / http.port.
func Target(domain *config.Domain) (string, error) {
	if domain == nil {
		return "", nil
	}

	if domain.HTTP == nil || domain.HTTP.Port == 0 {
		return domain.Base, nil
	}

	u, err := url.Parse(domain.Base)
	if err != nil {
		return "", fmt.Errorf("failed to parse base %q: %w", domain.Base, err)
	}

	if u.Port() != "" {
		return "", fmt.Errorf("base %q already includes a port; set it separately with --port instead", domain.Base)
	}

	u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(domain.HTTP.Port))

	return u.String(), nil
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
