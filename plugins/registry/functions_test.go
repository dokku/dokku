package registry

import "testing"

func TestNormalizeRegistryServer(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"hub.docker.com alias", "hub.docker.com", "docker.io"},
		{"docker.com alias", "docker.com", "docker.io"},
		{"docker.io untouched", "docker.io", "docker.io"},
		{"other registry untouched", "ghcr.io", "ghcr.io"},
		{"empty untouched", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeRegistryServer(tc.in); got != tc.want {
				t.Errorf("normalizeRegistryServer(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestConvertToHostname(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"bare hostname", "ghcr.io", "ghcr.io"},
		{"https scheme", "https://ghcr.io", "ghcr.io"},
		{"http scheme", "http://ghcr.io", "ghcr.io"},
		{"scheme and path", "https://index.docker.io/v1/", "index.docker.io"},
		{"trailing slash", "ghcr.io/", "ghcr.io"},
		{"port preserved", "localhost:5000", "localhost:5000"},
		{"port with scheme and path", "https://localhost:5000/v2/", "localhost:5000"},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToHostname(tc.in); got != tc.want {
				t.Errorf("convertToHostname(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestDockerAuthKey(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"docker.io", "docker.io", dockerIndexServer},
		{"hub.docker.com", "hub.docker.com", dockerIndexServer},
		{"docker.com", "docker.com", dockerIndexServer},
		{"index.docker.io", "index.docker.io", dockerIndexServer},
		{"index server address", dockerIndexServer, dockerIndexServer},
		{"empty server", "", dockerIndexServer},
		// docker does not treat registry-1.docker.io as the official index and
		// stores it under a key of its own
		{"registry-1.docker.io is not the index", "registry-1.docker.io", "registry-1.docker.io"},
		{"other registry", "ghcr.io", "ghcr.io"},
		{"other registry with scheme", "https://ghcr.io/v2/", "ghcr.io"},
		{"registry with port", "localhost:5000", "localhost:5000"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := dockerAuthKey(tc.in); got != tc.want {
				t.Errorf("dockerAuthKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
