package proxy

import (
	"testing"
)

func TestDecodeRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		want    Redirect
		wantErr bool
	}{
		{name: "with code", value: "example.com:302", want: Redirect{Source: "www.example.com", Destination: "example.com", Code: 302}},
		{name: "without code", value: "example.com", want: Redirect{Source: "www.example.com", Destination: "example.com", Code: DefaultRedirectCode}},
		{name: "empty destination", value: ":301", wantErr: true},
		{name: "invalid code", value: "example.com:abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := DecodeRedirect("www.example.com", tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRedirectEncodeRoundTrip(t *testing.T) {
	t.Parallel()

	redirect := Redirect{Source: "www.example.com", Destination: "example.com", Code: 307}
	got, err := DecodeRedirect(redirect.Source, redirect.Encode())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != redirect {
		t.Fatalf("got %+v, want %+v", got, redirect)
	}
}

func TestValidateRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		redirect Redirect
		wantErr  bool
	}{
		{name: "valid", redirect: Redirect{Source: "www.example.com", Destination: "example.com", Code: 301}},
		{name: "valid 308", redirect: Redirect{Source: "example.org", Destination: "example.com", Code: 308}},
		{name: "same hostname", redirect: Redirect{Source: "example.com", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "catch-all source", redirect: Redirect{Source: "_", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "wildcard source", redirect: Redirect{Source: "*.example.com", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "wildcard destination", redirect: Redirect{Source: "www.example.com", Destination: "*.example.com", Code: 301}, wantErr: true},
		{name: "source with scheme", redirect: Redirect{Source: "http://www.example.com", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "destination with path", redirect: Redirect{Source: "www.example.com", Destination: "example.com/path", Code: 301}, wantErr: true},
		{name: "destination with port", redirect: Redirect{Source: "www.example.com", Destination: "example.com:8080", Code: 301}, wantErr: true},
		{name: "leading dot", redirect: Redirect{Source: ".example.com", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "double dot", redirect: Redirect{Source: "www..example.com", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "empty source", redirect: Redirect{Source: "", Destination: "example.com", Code: 301}, wantErr: true},
		{name: "invalid code", redirect: Redirect{Source: "www.example.com", Destination: "example.com", Code: 200}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRedirect(tt.redirect)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %+v", tt.redirect)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %+v: %v", tt.redirect, err)
			}
		})
	}
}

func TestRedirectIsPermanent(t *testing.T) {
	t.Parallel()

	permanent := map[int]bool{301: true, 302: false, 303: false, 307: false, 308: true}
	for code, want := range permanent {
		if got := (Redirect{Code: code}).IsPermanent(); got != want {
			t.Fatalf("code %d: got %v, want %v", code, got, want)
		}
	}
}

func TestSortRedirects(t *testing.T) {
	t.Parallel()

	redirects := []Redirect{
		{Source: "www.example.org"},
		{Source: "example.net"},
		{Source: "www.example.com"},
	}
	SortRedirects(redirects)
	want := []string{"example.net", "www.example.com", "www.example.org"}
	for i, source := range want {
		if redirects[i].Source != source {
			t.Fatalf("position %d: got %s, want %s", i, redirects[i].Source, source)
		}
	}
}
