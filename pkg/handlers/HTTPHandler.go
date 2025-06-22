package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
)

func makeRequest(
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

	if domain != nil && domain.User != "" && domain.Pass != "" {
		req.SetBasicAuth(domain.User, domain.Pass)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func HTTPHandler(
	method string,
	domain *config.Domain,
	path string,
	reqBody *[]byte,
	headers map[string]string,
) ([]byte, error) {
	res, err := makeRequest(method, domain, path, reqBody, headers)
	if err != nil {
		fmt.Printf("Error making %s request: %v\n", method, err)
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
	}

	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	if len(resBody) == 0 {
		for key, values := range res.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
	}

	return resBody, nil
}

func setHeaders(req *http.Request) {
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
