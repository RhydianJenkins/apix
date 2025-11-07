package handlers

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
)

type mockResponse struct {
    StatusCode int
    Body string
    Headers map[string]string
}

func makeHandler (t *testing.T) http.HandlerFunc {
	mockResponse := mockResponse{
        StatusCode: http.StatusCreated,
        Body: `{"id": 123}`,
        Headers: map[string]string{
            "Content-Type": "application/json",
        },
    }

	handler := func (writer http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			t.Errorf("Expected POST, got %s", req.Method)
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf(
				"Wrong content type. Expected application/json but got %q",
				req.Header.Get("Content-Type"),
			)
		}
		if req.Header.Get("Accept") != "application/json" {
			t.Errorf(
				"Wrong 'Accept' header. Expected application/json but got %q",
				req.Header.Get("Accept"),
			)
		}
		if req.RequestURI != "/test" {
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

func TestPost(t *testing.T) {
	server := httptest.NewServer(makeHandler(t))
	defer server.Close()

	domain := &config.Domain{
		Base: server.URL,
		Name: "testapi",
		Pass: "",
		User: "",
	}

	path := "/test"

	body, err := HTTPHandler("POST", domain, path, nil, nil)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if body == nil {
		t.Errorf("Expected response body, got nil")
		return
	}

	if string(body) != `{"id": 123}` {
		t.Errorf("Unexpected body contents, got %q", string(body))
	}
}

func TestHeaderMerging(t *testing.T) {
	capturedHeaders := make(map[string][]string)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maps.Copy(capturedHeaders, r.Header)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	domain := &config.Domain{
		Base: server.URL,
		Name: "testapi",
		Headers: map[string]string{
			"Authorization": "Bearer config-token",
			"X-Config":      "config-value",
			"User-Agent":    "ConfigAgent/1.0",
		},
	}

	cliHeaders := map[string]string{
		"Authorization": "Bearer cli-token", // Should override config
		"X-CLI":         "cli-value",        // Should be added
		"Content-Type":  "text/plain",       // Should override default
	}

	_, err := HTTPHandler("GET", domain, "/test", nil, cliHeaders)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	tests := map[string]string{
		"Authorization": "Bearer cli-token", // CLI overrides config
		"X-Config":      "config-value",     // Config header preserved
		"X-Cli":         "cli-value",        // CLI header added (normalized to lowercase)
		"Content-Type":  "text/plain",       // CLI overrides default
		"Accept":        "application/json", // Default header preserved
		"User-Agent":    "ConfigAgent/1.0",  // Config overrides default
	}

	for expectedHeader, expectedValue := range tests {
		values, exists := capturedHeaders[expectedHeader]
		if !exists {
			t.Errorf("Expected header %q not found in request", expectedHeader)
			continue
		}

		if len(values) == 0 {
			t.Errorf("Header %q has no values", expectedHeader)
			continue
		}

		actualValue := values[0]
		if actualValue != expectedValue {
			t.Errorf("Header %q: expected %q, got %q", expectedHeader, expectedValue, actualValue)
		}
	}
}
