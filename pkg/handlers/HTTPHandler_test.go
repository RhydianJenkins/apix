package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
)

type MockResponse struct {
    StatusCode int
    Body string
    Headers map[string]string
}

func makeHandler (t *testing.T) http.HandlerFunc {
	mockResponse := MockResponse{
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

	body, err := HTTPHandler("POST", domain, nil, nil)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if body == nil {
		t.Errorf("Expected response body, got nil")
		return
	}

	if string(*body) != `{"id": 123}` {
		t.Errorf("Unexpected body contents, got %q", string(*body))
	}
}
