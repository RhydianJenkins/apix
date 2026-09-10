package handlers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestShowHandler(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	specPath := filepath.Join(tempDir, "spec.yaml")
	yamlContent := `openapi: 3.0.0
paths:
  /users/{id}:
    get:
      summary: Get a user by ID
      responses:
        "200":
          description: OK
`
	if err := os.WriteFile(specPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	config.SetDomain(&config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{
			OpenAPISpecPath: specPath,
		},
	})

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		ShowHandler(cmd, []string{"/users/{id}"})
	})

	if !strings.Contains(output, "GET /users/{id}") {
		t.Errorf("expected output to contain endpoint summary, got: %s", output)
	}

	if !strings.Contains(output, "Summary: Get a user by ID") {
		t.Errorf("expected output to contain summary, got: %s", output)
	}
}

func TestShowHandlerUnknownPath(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	specPath := filepath.Join(tempDir, "spec.yaml")
	yamlContent := `openapi: 3.0.0
paths:
  /users/{id}:
    get:
      summary: Get a user by ID
`
	if err := os.WriteFile(specPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	config.SetDomain(&config.Domain{
		Name: "testapi",
		Base: "https://api.example.com",
		HTTP: &config.HTTPOptions{
			OpenAPISpecPath: specPath,
		},
	})

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		ShowHandler(cmd, []string{"/does/not/exist"})
	})

	if !strings.Contains(output, "No endpoint found") {
		t.Errorf("expected a helpful 'no endpoint found' error, got: %s", output)
	}
}

func TestShowHandlerNoSpecConnected(t *testing.T) {
	tempDir := t.TempDir()
	config.CfgPath = filepath.Join(tempDir, ".apix.yaml")
	viper.Reset()

	config.SetDomain(&config.Domain{
		Name: "nospec",
		Base: "https://api.example.com",
	})

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		ShowHandler(cmd, []string{"/anything"})
	})

	if !strings.Contains(output, "No OpenAPI spec is connected") {
		t.Errorf("expected a helpful 'no OpenAPI spec connected' error, got: %s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()

	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
