package certs

import (
	"reflect"
	"testing"
)

func TestSubjectCommonName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"rfc2253 cn only", "subject=CN=dokku.me", "dokku.me"},
		{"rfc2253 multi rdn reversed", "subject=CN=node-js-app.dokku.me,OU=Operations,O=Expa,L=San Francisco,ST=California,C=US", "node-js-app.dokku.me"},
		{"rfc2253 wildcard", "subject=CN=*.dokku.me", "*.dokku.me"},
		{"openssl 3.x default spaced", "subject=CN = cn-only.example.com", "cn-only.example.com"},
		{"openssl 3.x default multi rdn", "subject=C=US, ST=California, L=San Francisco, O=Expa, OU=Operations, CN=node-js-app.dokku.me", "node-js-app.dokku.me"},
		{"legacy compat slash prefix", "subject=/CN=dokku.me", "dokku.me"},
		{"no subject prefix", "CN=dokku.me", "dokku.me"},
		{"no common name", "subject=OU=Operations,O=Expa,C=US", ""},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := subjectCommonName(tc.in); got != tc.want {
				t.Errorf("subjectCommonName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

const sanCertText = `Certificate:
    Data:
        X509v3 extensions:
            X509v3 Subject Alternative Name:
                DNS:www.test.dokku.me, DNS:www.test.app.dokku.me
`

func TestSSLHostnames(t *testing.T) {
	cases := []struct {
		name     string
		subject  string
		certText string
		want     []string
	}{
		{
			name:    "cn only no san",
			subject: "subject=CN=dokku.me",
			want:    []string{"dokku.me"},
		},
		{
			name:     "cn plus sans sorted",
			subject:  "subject=CN=test.dokku.me",
			certText: sanCertText,
			want:     []string{"test.dokku.me", "www.test.app.dokku.me", "www.test.dokku.me"},
		},
		{
			name:     "sans only no cn",
			subject:  "subject=OU=Operations",
			certText: sanCertText,
			want:     []string{"www.test.app.dokku.me", "www.test.dokku.me"},
		},
		{
			name:     "dedupes cn present in san",
			subject:  "subject=CN=www.test.dokku.me",
			certText: sanCertText,
			want:     []string{"www.test.app.dokku.me", "www.test.dokku.me"},
		},
		{
			name:    "no cn no san",
			subject: "subject=OU=Operations",
			want:    []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sslHostnames(tc.subject, tc.certText); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("sslHostnames(%q, ...) = %#v, want %#v", tc.subject, got, tc.want)
			}
		})
	}
}

func TestFormatSSLSubject(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"compat single", "subject=/CN=dokku.me", "CN=dokku.me"},
		{
			"compat multi rdn",
			"subject=/C=US/ST=California/L=San Francisco/O=Expa/OU=Operations/CN=node-js-app.dokku.me",
			"C=US; ST=California; L=San Francisco; O=Expa; OU=Operations; CN=node-js-app.dokku.me",
		},
		{
			"compat wildcard multi rdn",
			"subject=/OU=Domain Control Validated/OU=PositiveSSL Wildcard/CN=*.dokku.me",
			"OU=Domain Control Validated; OU=PositiveSSL Wildcard; CN=*.dokku.me",
		},
		{"legacy spaced prefix", "subject= /CN=dokku.me", "CN=dokku.me"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatSSLSubject(tc.in); got != tc.want {
				t.Errorf("formatSSLSubject(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
