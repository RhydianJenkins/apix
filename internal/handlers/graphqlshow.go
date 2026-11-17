package handlers

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/rhydianjenkins/apix/v3/internal/config"
	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
)

func GraphQLShowHandler(domain *config.Domain, name string, headers map[string]string) (string, error) {
	mergedHeaders := make(map[string]string)

	if domain != nil && domain.Headers != nil {
		maps.Copy(mergedHeaders, domain.Headers)
	}

	maps.Copy(mergedHeaders, headers)

	info, err := graphqlclient.ResolveOperation(domain, name, mergedHeaders)
	if err != nil {
		var statusErr *graphqlclient.StatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusUnauthorized && domain.HasAuth() {
			fmt.Fprintf(os.Stderr, "hint: run `apix auth %s` to refresh credentials\n", domain.Name)
		}
		return "", fmt.Errorf("failed to resolve operation %q: %w", name, err)
	}

	return graphqlclient.FormatOperationSummary(info), nil
}
