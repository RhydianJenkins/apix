package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
)

// reqBody is parsed as the operation's variables, keyed by arg name.
func GraphQLPostHandler(
	domain *config.Domain,
	name string,
	fieldPaths []string,
	reqBody *[]byte,
	headers map[string]string,
) ([]byte, error) {
	mergedHeaders := make(map[string]string)

	if domain != nil && domain.Headers != nil {
		maps.Copy(mergedHeaders, domain.Headers)
	}

	maps.Copy(mergedHeaders, headers)

	doc, err := graphqlclient.BuildQueryDocument(domain, name, fieldPaths, mergedHeaders)
	if err != nil {
		var statusErr *graphqlclient.StatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusUnauthorized && domain.HasAuth() {
			fmt.Fprintf(os.Stderr, "hint: run `apix auth %s` to refresh credentials\n", domain.Name)
		}
		return nil, fmt.Errorf("failed to build query for operation %q: %w", name, err)
	}

	variables := map[string]any{}
	if reqBody != nil && len(*reqBody) > 0 {
		if err := json.Unmarshal(*reqBody, &variables); err != nil {
			return nil, fmt.Errorf("failed to parse stdin as JSON variables: %w", err)
		}
	}

	payload, err := json.Marshal(map[string]any{
		"operationName": doc.Operation.Name,
		"query":         doc.Query,
		"variables":     variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build request body: %w", err)
	}

	return HTTPHandler(http.MethodPost, domain, "", &payload, headers)
}
