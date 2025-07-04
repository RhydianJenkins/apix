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
			name: "Run with absolute path",
			oasPath: filepath,
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

func makeTestTmpFile(name string, t *testing.T) (string) {
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
