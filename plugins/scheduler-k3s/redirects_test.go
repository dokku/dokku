package scheduler_k3s

import (
	"reflect"
	"testing"
)

func TestParseRedirects(t *testing.T) {
	redirects, err := parseRedirects("myapp", "old.example.com|app.example.com|308\nwww.example.com|app.example.com|302\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []ProcessRedirect{
		{Source: "old.example.com", Destination: "app.example.com", Code: 308, Permanent: true, Slug: "old-example-com"},
		{Source: "www.example.com", Destination: "app.example.com", Code: 302, Permanent: false, Slug: "www-example-com"},
	}
	if !reflect.DeepEqual(redirects, want) {
		t.Fatalf("got %#v, want %#v", redirects, want)
	}
}

func TestParseRedirectsEmpty(t *testing.T) {
	redirects, err := parseRedirects("myapp", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(redirects) != 0 {
		t.Fatalf("expected no redirects, got %#v", redirects)
	}
}

func TestParseRedirectsInvalid(t *testing.T) {
	for _, stdout := range []string{"www.example.com|app.example.com", "www.example.com|app.example.com|abc", "|app.example.com|301"} {
		if _, err := parseRedirects("myapp", stdout); err == nil {
			t.Fatalf("expected error for %q", stdout)
		}
	}
}

func TestFilterRedirectSources(t *testing.T) {
	redirects := []ProcessRedirect{{Source: "www.example.com"}}
	got := filterRedirectSources([]string{"app.example.com", "www.example.com"}, redirects)
	if !reflect.DeepEqual(got, []string{"app.example.com"}) {
		t.Fatalf("got %#v", got)
	}
}
