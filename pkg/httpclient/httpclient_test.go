package httpclient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/rhydianjenkins/apix/pkg/httpclient"
)

func TestDo_WithSeparatePort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL %q: %v", server.URL, err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("failed to parse test server port %q: %v", u.Port(), err)
	}

	domain := &config.Domain{
		Base: fmt.Sprintf("%s://%s", u.Scheme, u.Hostname()),
		HTTP: &config.HTTPOptions{Port: port},
	}

	resp, err := httpclient.Do("GET", domain, "/test", nil, nil)
	if err != nil {
		t.Fatalf("failed to make request with a separate port: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDo_PortInBase(t *testing.T) {
	domain := &config.Domain{
		Base: "http://localhost:1234",
		HTTP: &config.HTTPOptions{Port: 8080},
	}

	if _, err := httpclient.Do("GET", domain, "/test", nil, nil); err == nil {
		t.Fatal("expected an error when base includes a port")
	}
}
