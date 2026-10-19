package resource

import (
	"testing"
)

func TestAddMemorySuffixForDocker(t *testing.T) {
	testCases := []struct {
		key      string
		value    string
		expected string
	}{
		{key: "memory", value: "100", expected: "100m"},
		{key: "memory", value: "512m", expected: "512m"},
		{key: "memory", value: "512M", expected: "512M"},
		{key: "memory", value: "512MB", expected: "512MB"},
		{key: "memory", value: "1g", expected: "1g"},
		{key: "memory", value: "1G", expected: "1G"},
		{key: "memory", value: "1024k", expected: "1024k"},
		{key: "memory", value: "1024K", expected: "1024K"},
		{key: "memory", value: "1t", expected: "1t"},
		{key: "memory", value: "1048576b", expected: "1048576b"},
		{key: "cpu", value: "100", expected: "100"},
	}

	for _, tc := range testCases {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			actual := addMemorySuffixForDocker(tc.key, tc.value)
			if actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
