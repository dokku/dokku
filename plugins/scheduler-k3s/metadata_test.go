package scheduler_k3s

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseMetadataPairs(t *testing.T) {
	tests := []struct {
		name  string
		pairs []string
		want  map[string]string
	}{
		{
			name:  "empty list",
			pairs: []string{},
			want:  map[string]string{},
		},
		{
			name:  "multiple pairs",
			pairs: []string{"foo=bar", "baz=qux"},
			want:  map[string]string{"foo": "bar", "baz": "qux"},
		},
		{
			name:  "trailing equals stores an empty value",
			pairs: []string{"foo="},
			want:  map[string]string{"foo": ""},
		},
		{
			name:  "value containing equals is kept whole",
			pairs: []string{"foo=a=b"},
			want:  map[string]string{"foo": "a=b"},
		},
		{
			name:  "kubernetes style key",
			pairs: []string{"prometheus.io/scrape=true"},
			want:  map[string]string{"prometheus.io/scrape": "true"},
		},
		{
			name:  "multi-line value",
			pairs: []string{"foo=line one\nline two"},
			want:  map[string]string{"foo": "line one\nline two"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseMetadataPairs(test.pairs)
			if err != nil {
				t.Fatalf("parseMetadataPairs(%q) returned error: %s", test.pairs, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("parseMetadataPairs(%q) = %v, want %v", test.pairs, got, test.want)
			}
		})
	}
}

func TestParseMetadataPairsErrors(t *testing.T) {
	tests := []struct {
		name  string
		pairs []string
		want  string
	}{
		{
			name:  "missing equals",
			pairs: []string{"foo"},
			want:  "Invalid key=value pair: foo",
		},
		{
			name:  "empty key",
			pairs: []string{"=bar"},
			want:  "Invalid key=value pair: =bar",
		},
		{
			name:  "empty pair",
			pairs: []string{""},
			want:  "Invalid key=value pair: ",
		},
		{
			name:  "duplicate key",
			pairs: []string{"foo=bar", "foo=qux"},
			want:  "Duplicate key specified: foo",
		},
		{
			name:  "duplicate key with identical value",
			pairs: []string{"foo=bar", "foo=bar"},
			want:  "Duplicate key specified: foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseMetadataPairs(test.pairs)
			if err == nil {
				t.Fatalf("parseMetadataPairs(%q) = %v, want error", test.pairs, got)
			}
			if err.Error() != test.want {
				t.Errorf("parseMetadataPairs(%q) error = %q, want %q", test.pairs, err.Error(), test.want)
			}
		})
	}
}

func TestValidateResourceType(t *testing.T) {
	tests := []struct {
		name         string
		field        metadataField
		resourceType string
		wantErr      bool
	}{
		{
			name:         "valid annotation resource type",
			field:        annotationsField,
			resourceType: "deployment",
		},
		{
			name:         "valid label resource type",
			field:        labelsField,
			resourceType: "deployment",
		},
		{
			// An empty resource type is an absent filter for the clear commands.
			// The set commands reject it separately.
			name:         "empty resource type passes",
			field:        annotationsField,
			resourceType: "",
		},
		{
			name:         "unknown resource type",
			field:        annotationsField,
			resourceType: "deploymnet",
			wantErr:      true,
		},
		{
			// Keda resources do not flow user labels through, so they are
			// annotation-only resource types.
			name:         "keda resource type is valid for annotations",
			field:        annotationsField,
			resourceType: "keda_scaled_object",
		},
		{
			name:         "keda resource type is invalid for labels",
			field:        labelsField,
			resourceType: "keda_scaled_object",
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateResourceType(test.field, test.resourceType)
			if test.wantErr {
				if err == nil {
					t.Fatalf("validateResourceType(%s, %q) = nil, want error", test.field.Name, test.resourceType)
				}
				if !strings.Contains(err.Error(), "Invalid resource-type specified, valid resource types include: ") {
					t.Errorf("validateResourceType(%s, %q) error = %q, want the valid type list", test.field.Name, test.resourceType, err.Error())
				}
				for _, valid := range test.field.ResourceTypes {
					if !strings.Contains(err.Error(), valid) {
						t.Errorf("validateResourceType(%s, %q) error = %q, missing valid type %s", test.field.Name, test.resourceType, err.Error(), valid)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("validateResourceType(%s, %q) returned error: %s", test.field.Name, test.resourceType, err)
			}
		})
	}
}

func TestMetadataPropertyName(t *testing.T) {
	// Annotations are stored unprefixed and labels under a labels. prefix, which is
	// what getAnnotation and getLabel already build.
	if got := metadataPropertyName(annotationsField, "web", "deployment"); got != "web.deployment" {
		t.Errorf("metadataPropertyName(annotations, web, deployment) = %q, want web.deployment", got)
	}
	if got := metadataPropertyName(labelsField, "web", "deployment"); got != "labels.web.deployment" {
		t.Errorf("metadataPropertyName(labels, web, deployment) = %q, want labels.web.deployment", got)
	}
	if got := metadataPropertyName(annotationsField, GlobalProcessType, "pod"); got != "--global.pod" {
		t.Errorf("metadataPropertyName(annotations, --global, pod) = %q, want --global.pod", got)
	}
}
