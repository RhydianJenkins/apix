package oas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rhydianjenkins/apix/pkg/config"
	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	Paths map[string]any `json:"paths" yaml:"paths"`
}

// TODO Rhydian use something like this to parse OAS...
// https://pb33f.io/libopenapi/openapi/
func GetEndpointsValidArgs(specSource string) ([]string, error) {
	var data []byte
	var err error

	if specSource == "" {
		return make([]string, 0), nil
	}

	if strings.HasPrefix(specSource, "http://") || strings.HasPrefix(specSource, "https://") {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(specSource)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = os.ReadFile(specSource)
		if err != nil {
			return nil, err
		}
	}

	var spec OpenAPISpec
	err = json.Unmarshal(data, &spec)
	if err != nil {
		err = yaml.Unmarshal(data, &spec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: %w", err)
		}
	}

	var endpoints []string
	for path := range spec.Paths {
		endpoints = append(endpoints, path)
	}

	return endpoints, nil
}

func HasValidOpenAPISpec(d *config.Domain) bool {
    if d.OpenAPISpecPath == "" {
        return false
    }

    if strings.HasPrefix(d.OpenAPISpecPath, "http://") || strings.HasPrefix(d.OpenAPISpecPath, "https://") {
        return true
    }

    _, err := os.Stat(d.OpenAPISpecPath)
    return err == nil
}
