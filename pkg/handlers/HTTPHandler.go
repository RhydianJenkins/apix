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
	reqBody *[]byte,
	headers map[string]string,
) (*http.Response, *[]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var req *http.Request
	var err error

	url := domain.Base

	if reqBody != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(*reqBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, &responseBody, nil
}

func HTTPHandler(
	method string,
	domain *config.Domain,
	reqBody *[]byte,
	headers map[string]string,
) (*[]byte, error) {
	res, resBody, err := makeRequest(method, domain, reqBody, headers)

	if err != nil {
		fmt.Printf("Error making %s request: %v\n", method, err)
	}

	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	if len(*resBody) == 0 {
		fmt.Printf("Code: %d. Body: empty. Headers:\n", res.StatusCode)
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

	if req.Header.Get("accept") == "" {
		req.Header.Set("Accept", "*/*")
	}

	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	}
}
