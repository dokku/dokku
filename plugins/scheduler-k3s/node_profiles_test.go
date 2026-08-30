package scheduler_k3s

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chartutil"
)

func TestValidateNodeProfileName(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		wantErr     bool
	}{
		{name: "typical name", profileName: "edge-workers"},
		{name: "single character", profileName: "a"},
		{name: "digits", profileName: "pool0"},
		{name: "maximum length", profileName: strings.Repeat("a", maxNodeProfileNameLength)},
		{name: "empty", profileName: "", wantErr: true},
		{name: "uppercase", profileName: "EdgePool", wantErr: true},
		{name: "one over the maximum length", profileName: strings.Repeat("a", maxNodeProfileNameLength+1), wantErr: true},
		{name: "twenty seven characters", profileName: "a-twenty-seven-char-profile", wantErr: true},
		{name: "leading dash", profileName: "-edge", wantErr: true},
		{name: "trailing dash", profileName: "edge-", wantErr: true},
		{name: "dot", profileName: "edge.workers", wantErr: true},
		{name: "underscore", profileName: "edge_workers", wantErr: true},
		{name: "space", profileName: "edge workers", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateNodeProfileName(test.profileName)
			if test.wantErr && err == nil {
				t.Errorf("validateNodeProfileName(%q) = nil, want an error", test.profileName)
			}
			if !test.wantErr && err != nil {
				t.Errorf("validateNodeProfileName(%q) = %v, want nil", test.profileName, err)
			}
		})
	}
}

// TestMaxNodeProfileNameLengthMatchesHelm asserts the limit validateNodeProfileName
// reports is the same one helm enforces. Without it the local arithmetic could drift
// from chartutil.ValidateReleaseName, which is the failure this limit exists to prevent.
func TestMaxNodeProfileNameLengthMatchesHelm(t *testing.T) {
	longest := getNodeSysctlsReleaseName(strings.Repeat("a", maxNodeProfileNameLength))
	if err := chartutil.ValidateReleaseName(longest); err != nil {
		t.Errorf("chartutil.ValidateReleaseName(%q) = %v, want nil", longest, err)
	}

	tooLong := getNodeSysctlsReleaseName(strings.Repeat("a", maxNodeProfileNameLength+1))
	if err := chartutil.ValidateReleaseName(tooLong); err == nil {
		t.Errorf("chartutil.ValidateReleaseName(%q) = nil, want an error", tooLong)
	}
}

// TestValidateStoredNodeProfileNameAcceptsLegacyNames asserts profiles:remove can still
// address a profile stored before names were validated against their helm release name
func TestValidateStoredNodeProfileNameAcceptsLegacyNames(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		wantErr     bool
	}{
		{name: "uppercase", profileName: "EdgePool"},
		{name: "twenty seven characters", profileName: "a-twenty-seven-char-profile"},
		{name: "legacy maximum length", profileName: strings.Repeat("a", legacyMaxNodeProfileNameLength)},
		{name: "empty", profileName: "", wantErr: true},
		{name: "one over the legacy maximum length", profileName: strings.Repeat("a", legacyMaxNodeProfileNameLength+1), wantErr: true},
		{name: "underscore", profileName: "edge_workers", wantErr: true},
		{name: "path traversal", profileName: "../config", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateStoredNodeProfileName(test.profileName)
			if test.wantErr && err == nil {
				t.Errorf("validateStoredNodeProfileName(%q) = nil, want an error", test.profileName)
			}
			if !test.wantErr && err != nil {
				t.Errorf("validateStoredNodeProfileName(%q) = %v, want nil", test.profileName, err)
			}
		})
	}
}

// TestVerifyNodeProfileSysctlsSupported asserts node-sysctls:set refuses a legacy profile
// name rather than storing sysctls the reconcile loop would have to skip
func TestVerifyNodeProfileSysctlsSupported(t *testing.T) {
	if err := verifyNodeProfileSysctlsSupported("edge-workers"); err != nil {
		t.Errorf("verifyNodeProfileSysctlsSupported(\"edge-workers\") = %v, want nil", err)
	}

	if err := verifyNodeProfileSysctlsSupported("EdgePool"); err == nil {
		t.Error("verifyNodeProfileSysctlsSupported(\"EdgePool\") = nil, want an error")
	}
}
