package resource

import (
	"testing"
)

func TestFormatResourceValueForDocker(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    string
		expected string
		wantErr  bool
	}{
		{name: "unitless memory defaults to megabytes", key: "memory", value: "100", expected: "100m"},
		{name: "unitless decimal memory defaults to megabytes", key: "memory", value: "1.5", expected: "1.5m"},
		{name: "lowercase memory unit", key: "memory", value: "512m", expected: "512m"},
		{name: "uppercase megabyte unit", key: "memory", value: "512M", expected: "512M"},
		{name: "uppercase gigabyte unit", key: "memory", value: "1G", expected: "1G"},
		{name: "uppercase kilobyte unit", key: "memory", value: "1024K", expected: "1024K"},
		{name: "terabyte unit", key: "memory", value: "1t", expected: "1t"},
		{name: "two letter unit", key: "memory", value: "512MB", expected: "512MB"},
		{name: "binary unit", key: "memory", value: "1GiB", expected: "1GiB"},
		{name: "byte unit", key: "memory", value: "100b", expected: "100b"},
		{name: "memory rejects negative values", key: "memory", value: "-1", wantErr: true},
		{name: "memory rejects extra arguments", key: "memory", value: "100G --foo", wantErr: true},
		{name: "memory rejects invalid units", key: "memory", value: "1Gm", wantErr: true},
		{name: "unlimited memory-swap", key: "memory-swap", value: "-1", expected: "-1"},
		{name: "unitless memory-swap stays in bytes", key: "memory-swap", value: "512", expected: "512"},
		{name: "uppercase memory-swap unit", key: "memory-swap", value: "2G", expected: "2G"},
		{name: "memory-swap rejects invalid units", key: "memory-swap", value: "1x", wantErr: true},
		{name: "memory-swap rejects other negative values", key: "memory-swap", value: "-2", wantErr: true},
		{name: "other keys are passed through", key: "cpus", value: "1.5", expected: "1.5"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, err := formatResourceValueForDocker(testCase.key, testCase.value)
			if testCase.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s=%s, got %s", testCase.key, testCase.value, value)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %s=%s: %v", testCase.key, testCase.value, err)
			}

			if value != testCase.expected {
				t.Fatalf("expected %s, got %s", testCase.expected, value)
			}
		})
	}
}

func TestIsResourceTypeKey(t *testing.T) {
	testCases := []struct {
		name         string
		key          string
		resourceType string
		expected     bool
	}{
		{name: "default limit key", key: "_default_.limit.memory", resourceType: "limit", expected: true},
		{name: "process limit key", key: "web.limit.cpu", resourceType: "limit", expected: true},
		{name: "process reserve key", key: "web.reserve.memory", resourceType: "reserve", expected: true},
		{name: "dotted process type", key: "foo.bar.limit.memory", resourceType: "limit", expected: true},
		{name: "reserve key does not match limit", key: "_default_.reserve.memory", resourceType: "limit", expected: false},
		{name: "limit key does not match reserve", key: "web.limit.memory", resourceType: "reserve", expected: false},
		{name: "key without process type", key: "limit.memory", resourceType: "limit", expected: false},
		{name: "key with a single part", key: "limit", resourceType: "limit", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value := isResourceTypeKey(testCase.key, testCase.resourceType)
			if value != testCase.expected {
				t.Fatalf("expected %v for %s/%s, got %v", testCase.expected, testCase.key, testCase.resourceType, value)
			}
		})
	}
}
