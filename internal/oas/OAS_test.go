package oas

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type mockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

type testCase struct {
	name     string
	oasPath  string
	expected []string
}

func TestGetEndpointsValidArgs(t *testing.T) {
	server := httptest.NewServer(makeHandler(t))
	defer server.Close()
	filename := "webhooks.yaml"
	filepath := makeTestTmpFile(filename, t)

	tests := []testCase{
		{
			name:     "Run with https url",
			oasPath:  server.URL + "/" + filename,
			expected: []string{"/webhooks/subscription/{id}"},
		},
		{
			name:     "Run with absolute path",
			oasPath:  filepath,
			expected: []string{"/webhooks/subscription/{id}"},
		},
	}

	for _, test := range tests {
		t.Run(makeTest(test))
	}
}

func makeTest(test testCase) (string, func(t *testing.T)) {
	return test.name, func(t *testing.T) {
		t.Log(test.name)

		got, err := GetEndpointsValidArgs("get", test.oasPath)

		if err != nil {
			t.Errorf("GetEndpointsValidArgs() returned error %v", err)
			return
		}

		if got == nil {
			t.Errorf("GetEndpointsValidArgs() = nil, expected %v", test.expected)
			return
		}

		if len(got) != len(test.expected) {
			t.Errorf("GetEndpointsValidArgs() = %v, expected %v", got, test.expected)
			return
		}

		for i, v := range got {
			if v != test.expected[i] {
				t.Errorf("GetEndpointsValidArgs() = %v, expected %v", got, test.expected)
				return
			}
		}
	}
}

func makeHandler(t *testing.T) http.HandlerFunc {
	mockResponse := mockResponse{
		StatusCode: http.StatusCreated,
		Body: `{
			"openapi": 3.1.0,
			"paths": {
				"/webhooks/subscription/{id}": {
					"get": {}
				},
				"/webhooks/subscription": {
					"post": {},
				}
			}
		}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	handler := func(writer http.ResponseWriter, req *http.Request) {
		if req.RequestURI != "/webhooks.yaml" {
			t.Errorf(
				"Wrong request URI. Expected /test but got %q",
				req.RequestURI,
			)
		}

		writer.WriteHeader(mockResponse.StatusCode)
		for key, value := range mockResponse.Headers {
			writer.Header().Set(key, value)
		}
		writer.Write([]byte(mockResponse.Body))
	}

	return http.HandlerFunc(handler)
}

func TestGetPathItem(t *testing.T) {
	filename := "webhooks.yaml"
	filepath := makeTestTmpFile(filename, t)

	t.Run("returns the path item when the path exists", func(t *testing.T) {
		pathItem, err := GetPathItem("/webhooks/subscription/{id}", filepath)

		if err != nil {
			t.Fatalf("GetPathItem() returned error %v", err)
		}

		if pathItem == nil {
			t.Fatalf("GetPathItem() = nil, expected a path item")
		}

		if pathItem.Get == nil {
			t.Errorf("expected GET operation to be present on path item")
		}
	})

	t.Run("returns an error when the path does not exist", func(t *testing.T) {
		_, err := GetPathItem("/does/not/exist", filepath)

		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})

	t.Run("returns an error when no spec is connected", func(t *testing.T) {
		_, err := GetPathItem("/webhooks/subscription/{id}", "")

		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})
}

func TestListEndpoints(t *testing.T) {
	filename := "endpoints.yaml"
	path := makeTestTmpFile(filename, t)

	t.Run("returns nil when no spec is connected", func(t *testing.T) {
		got, err := ListEndpoints("", "")
		if err != nil {
			t.Fatalf("ListEndpoints() returned error %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("lists every operation across every path when unfiltered", func(t *testing.T) {
		got, err := ListEndpoints("", path)
		if err != nil {
			t.Fatalf("ListEndpoints() returned error %v", err)
		}

		want := []EndpointInfo{
			{Method: "POST", Path: "/webhooks/subscription"},
			{Method: "GET", Path: "/webhooks/subscription/{id}"},
		}

		if len(got) != len(want) {
			t.Fatalf("expected %d endpoints, got %d: %v", len(want), len(got), got)
		}

		for i, w := range want {
			if got[i].Method != w.Method || got[i].Path != w.Path {
				t.Errorf("endpoint %d: expected %+v, got %+v", i, w, got[i])
			}
		}
	})

	t.Run("filters by method case-insensitively", func(t *testing.T) {
		got, err := ListEndpoints("post", path)
		if err != nil {
			t.Fatalf("ListEndpoints() returned error %v", err)
		}

		if len(got) != 1 || got[0].Path != "/webhooks/subscription" || got[0].Method != "POST" {
			t.Errorf("expected only the POST /webhooks/subscription endpoint, got %v", got)
		}
	})

	t.Run("returns an error for an invalid spec source", func(t *testing.T) {
		_, err := ListEndpoints("", "/does/not/exist.yaml")
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func makeTestTmpFile(name string, t *testing.T) string {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, name)

	yamlContent := `openapi: 3.1.0
paths:
  /webhooks/subscription/{id}:
    get: {}
  /webhooks/subscription:
    post: {}
`

	err := os.WriteFile(tempFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tempFile
}
