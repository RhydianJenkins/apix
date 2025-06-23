package oas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockResponse struct {
    StatusCode int
    Body string
    Headers map[string]string
}

type testCase struct {
	name string
	oasPath string
	expected []string
}

func makeHandler (t *testing.T) http.HandlerFunc {
	mockResponse := mockResponse{
        StatusCode: http.StatusCreated,
		Body: `{ "paths": {
			"/webhooks/subscription": {
				"get": {},
			},
			"/webhooks/subscription/{id}": {
				"get": {}
			}
		}}`,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
    }

	handler := func (writer http.ResponseWriter, req *http.Request) {
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

func TestGetEndpointsValidArgs(t *testing.T) {
	server := httptest.NewServer(makeHandler(t))
	defer server.Close()

	tests := []testCase{
		{
			name: "Run with https url",
			oasPath: server.URL + "/webhooks.yaml",
			expected: []string{"/webhooks/subscription", "/webhooks/subscription/{id}"},
		},
		{
			name: "Run with absolute path",
			oasPath: "/home/rhydian/code/basekit/openapi-specification/webhooks.yaml",
			expected: []string{"/webhooks/subscription", "/webhooks/subscription/{id}"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)

			if strings.HasPrefix(test.oasPath, "http") {
			}

			got, err := GetEndpointsValidArgs(test.oasPath)

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
		})
	}
}

