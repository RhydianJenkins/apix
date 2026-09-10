package handlers

import (
	"fmt"
	"io"
	"maps"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/httpclient"
)

func HTTPHandler(
	method string,
	domain *config.Domain,
	path string,
	reqBody *[]byte,
	headers map[string]string,
) ([]byte, error) {
	mergedHeaders := make(map[string]string)

	if domain != nil && domain.Headers != nil {
		maps.Copy(mergedHeaders, domain.Headers)
	}

	maps.Copy(mergedHeaders, headers)

	res, err := httpclient.Do(method, domain, path, reqBody, mergedHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to make %s request: %w", method, err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
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
