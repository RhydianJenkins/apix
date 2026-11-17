package graphqlclient_test

import (
	"strings"
	"testing"

	"github.com/rhydianjenkins/apix/v3/internal/graphqlclient"
)

func TestFormatOperationSummary(t *testing.T) {
	info := &graphqlclient.OperationInfo{
		Kind:        "Query",
		Name:        "user",
		Description: "Fetch a user by ID.",
		ReturnType:  "User",
		Args: []graphqlclient.ArgInfo{
			{Name: "id", Type: "ID!", Description: "The user's ID."},
		},
	}

	out := graphqlclient.FormatOperationSummary(info)

	wantContains := []string{
		"Query.user(id: ID!) -> User",
		"Description: Fetch a user by ID.",
		"id: ID! - The user's ID.",
	}

	for _, want := range wantContains {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}

	if strings.Contains(out, "Deprecated") {
		t.Errorf("expected no deprecation notice for a non-deprecated operation, got:\n%s", out)
	}
}

func TestFormatOperationSummary_ExpandsInputObjectAndEnumArgs(t *testing.T) {
	info := &graphqlclient.OperationInfo{
		Kind:       "Query",
		Name:       "account",
		ReturnType: "Account",
		Args: []graphqlclient.ArgInfo{
			{
				Name: "identifier",
				Type: "AccountIdentifier!",
				InputFields: []graphqlclient.ArgInfo{
					{Name: "id", Type: "String", Description: "Account UUID."},
					{Name: "number", Type: "String"},
				},
			},
			{
				Name:       "state",
				Type:       "AccountState",
				EnumValues: []string{"ACTIVE", "CLOSED"},
			},
		},
	}

	out := graphqlclient.FormatOperationSummary(info)

	wantContains := []string{
		"identifier: AccountIdentifier!",
		"Fields:",
		"id: String - Account UUID.",
		"number: String",
		"state: AccountState",
		"values: ACTIVE, CLOSED",
	}

	for _, want := range wantContains {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestFormatOperationSummary_RecursiveRefNotedInsteadOfLooping(t *testing.T) {
	info := &graphqlclient.OperationInfo{
		Kind:       "Query",
		Name:       "account",
		ReturnType: "Account",
		Args: []graphqlclient.ArgInfo{
			{
				Name: "self",
				Type: "SelfRefInput",
				InputFields: []graphqlclient.ArgInfo{
					{Name: "nested", Type: "SelfRefInput", IsRecursiveRef: true},
				},
			},
		},
	}

	out := graphqlclient.FormatOperationSummary(info)

	if !strings.Contains(out, "recursive reference to SelfRefInput") {
		t.Errorf("expected a recursive reference note, got:\n%s", out)
	}
}

func TestFormatOperationSummary_Deprecated(t *testing.T) {
	info := &graphqlclient.OperationInfo{
		Kind:              "Query",
		Name:              "search",
		ReturnType:        "[String!]!",
		IsDeprecated:      true,
		DeprecationReason: "Use user search instead.",
	}

	out := graphqlclient.FormatOperationSummary(info)

	if !strings.Contains(out, "Query.search() -> [String!]!") {
		t.Errorf("expected signature with no args, got:\n%s", out)
	}

	if !strings.Contains(out, "Deprecated: Use user search instead.") {
		t.Errorf("expected deprecation reason, got:\n%s", out)
	}
}
