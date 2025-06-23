package oas_test

import(
	"github.com/rhydianjenkins/apix/pkg/oas"
	"testing"
)

type testCase struct {
	name string
	oasPath string
	expected []string
}

func TestGetEndpointsValidArgs(t *testing.T) {
	tests := []testCase{
		// {
		// 	name: "Run with https url",
		// 	oasPath: "https://todo.example.com/webhooks.yaml",
		// 	expected: []string{"/webhooks/subscription", "/webhooks/subscription/{id}"},
		// },
		{
			name: "Run with absolute path",
			oasPath: "/home/rhydian/code/basekit/openapi-specification/webhooks.yaml",
			expected: []string{"/webhooks/subscription", "/webhooks/subscription/{id}"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)

			got, err := oas.GetEndpointsValidArgs(test.oasPath)

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

