package oas

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/rhydianjenkins/apix/pkg/config"
)

func GetEndpointsValidArgs(method, specSource string) ([]string, error) {
	if specSource == "" {
		return make([]string, 0), nil
	}

	oasDocument, err := loadDocument(specSource)

	if err != nil {
		return nil, err
	}

	v3Model, errors := oasDocument.BuildV3Model()

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to build OpenAPI v3 model: %v", errors)
	}

	endpoints := []string{}
	pathItems := v3Model.Model.Paths.PathItems

	for path, pathItem := range pathItems.FromNewest() {
		if method == "" || hasMethod(pathItem, method) {
			endpoints = append(endpoints, path)
		}
	}

	return endpoints, nil
}

// GetPathItem loads the given OpenAPI spec and returns the PathItem describing
// the operations available at the given path. It returns a descriptive error
// if no spec is connected, the spec fails to load/parse, or the path does not
// exist in the spec.
func GetPathItem(path, specSource string) (*v3.PathItem, error) {
	if specSource == "" {
		return nil, fmt.Errorf("No OpenAPI spec is connected to the active domain.\nConnect one with `apix new --oas <path/url>` or `apix edit`.\n")
	}

	oasDocument, err := loadDocument(specSource)

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	v3Model, errors := oasDocument.BuildV3Model()

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to build OpenAPI v3 model: %v", errors)
	}

	pathItem := v3Model.Model.Paths.PathItems.GetOrZero(path)

	if pathItem == nil {
		return nil, fmt.Errorf("No endpoint found for path %q in the connected OpenAPI spec.\n", path)
	}

	return pathItem, nil
}

func loadDocument(specSource string) (libopenapi.Document, error) {
	if strings.HasPrefix(specSource, "http://") || strings.HasPrefix(specSource, "https://") {
		return loadFromRemoteUrl(specSource)
	}

	return loadFromLocalPath(specSource)
}

func loadFromRemoteUrl(remoteUrl string) (libopenapi.Document, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(remoteUrl)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	baseUrl, err := url.Parse(remoteUrl)

	if err != nil {
		return nil, err
	}

	config := &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BaseURL:               baseUrl,
	}

	oasData, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return libopenapi.NewDocumentWithConfiguration(oasData, config)
}

func loadFromLocalPath(basePath string) (libopenapi.Document, error) {
	oasData, err := os.ReadFile(basePath)

	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(basePath)
	config := &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		AllowRemoteReferences: true,
		BasePath:              baseDir,
	}

	return libopenapi.NewDocumentWithConfiguration(oasData, config)
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

func hasMethod(pathItem *v3.PathItem, method string) bool {
	switch strings.ToUpper(method) {
	case "GET":
		return pathItem.Get != nil
	case "POST":
		return pathItem.Post != nil
	case "PUT":
		return pathItem.Put != nil
	case "DELETE":
		return pathItem.Delete != nil
	case "PATCH":
		return pathItem.Patch != nil
	case "HEAD":
		return pathItem.Head != nil
	case "OPTIONS":
		return pathItem.Options != nil
	default:
		return false
	}
}
