package oas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatPathItemSummary(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "spec.yaml")

	yamlContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get a user by ID
      description: Returns a single user record.
      operationId: getUserById
      tags:
        - users
      parameters:
        - name: id
          in: path
          required: true
          description: The user ID
          schema:
            type: string
      responses:
        "200":
          description: Successful response
        "404":
          description: User not found
    delete:
      summary: Delete a user
      deprecated: true
      responses:
        "204":
          description: No content
  /users:
    post:
      summary: Create a user
      requestBody:
        required: true
        description: The user to create
        content:
          application/json:
            schema:
              type: object
      responses:
        "201":
          description: Created
`

	if err := os.WriteFile(specPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp spec file: %v", err)
	}

	t.Run("formats an endpoint with parameters and responses", func(t *testing.T) {
		pathItem, err := GetPathItem("/users/{id}", specPath)
		if err != nil {
			t.Fatalf("GetPathItem() returned error %v", err)
		}

		got := FormatPathItemSummary("/users/{id}", pathItem)

		expectedSubstrings := []string{
			"GET /users/{id}",
			"Summary: Get a user by ID",
			"Description: Returns a single user record.",
			"Operation ID: getUserById",
			"Tags: users",
			"id (path, required) string - The user ID",
			"200: Successful response",
			"404: User not found",
			"DELETE /users/{id}",
			"Deprecated: true",
			"204: No content",
		}

		for _, expected := range expectedSubstrings {
			if !strings.Contains(got, expected) {
				t.Errorf("expected output to contain %q, got:\n%s", expected, got)
			}
		}
	})

	t.Run("formats a request body", func(t *testing.T) {
		pathItem, err := GetPathItem("/users", specPath)
		if err != nil {
			t.Fatalf("GetPathItem() returned error %v", err)
		}

		got := FormatPathItemSummary("/users", pathItem)

		expectedSubstrings := []string{
			"POST /users",
			"Required: true",
			"Description: The user to create",
			"Content-Types: application/json",
			"201: Created",
		}

		for _, expected := range expectedSubstrings {
			if !strings.Contains(got, expected) {
				t.Errorf("expected output to contain %q, got:\n%s", expected, got)
			}
		}
	})
}
