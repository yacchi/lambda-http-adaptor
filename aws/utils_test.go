package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExpandPathParameters(t *testing.T) {
	cases := []struct {
		resourcePath   string
		pathParameters map[string]string
		expected       string
	}{
		{
			resourcePath: "/{proxy+}",
			pathParameters: map[string]string{
				"proxy": "api/v1/test",
			},
			expected: "/api/v1/test",
		},
		{
			resourcePath: "/api/{proxy+}",
			pathParameters: map[string]string{
				"proxy": "v1/test",
			},
			expected: "/api/v1/test",
		},
		{
			resourcePath: "/api/{version}/{action}",
			pathParameters: map[string]string{
				"version": "v1",
				"action":  "test",
			},
			expected: "/api/v1/test",
		},
		{
			resourcePath: "/api/{version}/{action}/other",
			pathParameters: map[string]string{
				"version": "v1",
				"action":  "test",
			},
			expected: "/api/v1/test/other",
		},
		{
			resourcePath: "/api/{version}/{action}/{unknown}",
			pathParameters: map[string]string{
				"version": "v1",
				"action":  "test",
			},
			expected: "/api/v1/test/{unknown}",
		},
		{
			resourcePath: "/api/{version}/{action}/{brace_start",
			pathParameters: map[string]string{
				"version": "v1",
				"action":  "test",
			},
			expected: "/api/v1/test/{brace_start",
		},
	}

	for _, c := range cases {
		actual := ExpandPathParameters(c.resourcePath, c.pathParameters)
		if c.expected != actual {
			assert.Equal(t, c.expected, actual)
		}
	}
}
