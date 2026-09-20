package registry

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func encodeAuth(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func hubConfig() string {
	return fmt.Sprintf(`{"auths":{"https://index.docker.io/v1/":{"auth":%q}}}`, encodeAuth("user:pass"))
}

func TestCheckAuthStatus(t *testing.T) {
	cases := []struct {
		name     string
		config   string
		server   string
		username string
		password string
		want     int
	}{
		{"no config with credential", "", "docker.io", "user", "pass", authStatusMissing},
		{"no config without credential", "", "docker.io", "", "", authStatusMatch},
		{"empty auths after logout", `{"auths":{}}`, "docker.io", "user", "pass", authStatusMissing},
		{"empty auths without credential", `{"auths":{}}`, "docker.io", "", "", authStatusMatch},

		{"hub via docker.io", hubConfig(), "docker.io", "user", "pass", authStatusMatch},
		{"hub via hub.docker.com", hubConfig(), "hub.docker.com", "user", "pass", authStatusMatch},
		{"hub via docker.com", hubConfig(), "docker.com", "user", "pass", authStatusMatch},
		{"hub via index.docker.io", hubConfig(), "index.docker.io", "user", "pass", authStatusMatch},
		{"hub via index server address", hubConfig(), dockerIndexServer, "user", "pass", authStatusMatch},
		// docker stores registry-1.docker.io under its own key, so a hub
		// credential must not be reported as covering it
		{"registry-1.docker.io is not hub", hubConfig(), "registry-1.docker.io", "user", "pass", authStatusMissing},
		{"wrong password", hubConfig(), "docker.io", "user", "wrong", authStatusDiffers},
		{"wrong username", hubConfig(), "docker.io", "other", "pass", authStatusDiffers},
		{"no username with credential stored", hubConfig(), "docker.io", "", "", authStatusDiffers},
		{"unrelated server", hubConfig(), "ghcr.io", "user", "pass", authStatusMissing},

		{"other registry exact key", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q}}}`, encodeAuth("user:pass")), "ghcr.io", "user", "pass", authStatusMatch},
		{"other registry legacy key", fmt.Sprintf(`{"auths":{"https://ghcr.io/v1/":{"auth":%q}}}`, encodeAuth("user:pass")), "ghcr.io", "user", "pass", authStatusMatch},
		{"registry with port", fmt.Sprintf(`{"auths":{"localhost:5000":{"auth":%q}}}`, encodeAuth("user:pass")), "localhost:5000", "user", "pass", authStatusMatch},
		{"password containing a colon", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q}}}`, encodeAuth("user:pa:ss")), "ghcr.io", "user", "pa:ss", authStatusMatch},
		{"plaintext username and password fields", `{"auths":{"ghcr.io":{"username":"user","password":"pass"}}}`, "ghcr.io", "user", "pass", authStatusMatch},

		{"identity token with matching username", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q,"identitytoken":"token"}}}`, encodeAuth("user:")), "ghcr.io", "user", "pass", authStatusUnknown},
		{"identity token with differing username", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q,"identitytoken":"token"}}}`, encodeAuth("user:")), "ghcr.io", "other", "pass", authStatusDiffers},
		{"identity token without credential", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q,"identitytoken":"token"}}}`, encodeAuth("user:")), "ghcr.io", "", "", authStatusDiffers},

		{"credsStore holds the secret", `{"auths":{"ghcr.io":{}},"credsStore":"secretservice"}`, "ghcr.io", "user", "pass", authStatusUnknown},
		{"credsStore without credential", `{"auths":{"ghcr.io":{}},"credsStore":"secretservice"}`, "ghcr.io", "", "", authStatusDiffers},
		{"credHelpers for this server", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q}},"credHelpers":{"ghcr.io":"ecr-login"}}`, encodeAuth("user:pass")), "ghcr.io", "user", "pass", authStatusUnknown},
		{"credHelpers for another server", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q}},"credHelpers":{"other.io":"ecr-login"}}`, encodeAuth("user:pass")), "ghcr.io", "user", "pass", authStatusMatch},
		{"credHelpers keyed by hub index server", `{"auths":{"https://index.docker.io/v1/":{}},"credHelpers":{"https://index.docker.io/v1/":"desktop"}}`, "docker.io", "user", "pass", authStatusUnknown},

		{"empty entry", `{"auths":{"ghcr.io":{}}}`, "ghcr.io", "user", "pass", authStatusUnknown},
		{"undecodable auth", `{"auths":{"ghcr.io":{"auth":"!!!not base64!!!"}}}`, "ghcr.io", "user", "pass", authStatusUnknown},
		{"auth without a separator", fmt.Sprintf(`{"auths":{"ghcr.io":{"auth":%q}}}`, encodeAuth("nocolon")), "ghcr.io", "user", "pass", authStatusUnknown},
		{"invalid json", `{`, "ghcr.io", "user", "pass", authStatusUnknown},
		{"auths is not an object", `{"auths":"nope"}`, "ghcr.io", "user", "pass", authStatusUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var contents []byte
			if tc.config != "" {
				contents = []byte(tc.config)
			}

			got, message := checkAuthStatus(authStatusInput{
				ConfigBytes: contents,
				Server:      tc.server,
				Username:    tc.username,
				Password:    tc.password,
			})
			if got != tc.want {
				t.Errorf("checkAuthStatus() = %d, want %d", got, tc.want)
			}
			if got == authStatusUnknown && message == "" {
				t.Error("checkAuthStatus() returned no reason for an unreadable credential")
			}
			if got != authStatusUnknown && message != "" {
				t.Errorf("checkAuthStatus() returned an unexpected message: %s", message)
			}
		})
	}
}

func TestCheckAuthStatusDoesNotLeakCredentials(t *testing.T) {
	_, message := checkAuthStatus(authStatusInput{
		ConfigBytes: []byte(`{"auths":{"ghcr.io":{}},"credsStore":"secretservice"}`),
		Server:      "ghcr.io",
		Username:    "user",
		Password:    "supersecret",
	})
	if message == "" {
		t.Fatal("expected a reason to be reported")
	}
	for _, secret := range []string{"supersecret", "user"} {
		if strings.Contains(message, secret) {
			t.Errorf("message %q leaks %q", message, secret)
		}
	}
}

func TestAuthServersFromConfig(t *testing.T) {
	cases := []struct {
		name   string
		config string
		want   []string
	}{
		{"no config", "", []string{}},
		{"no credentials", `{"auths":{}}`, []string{}},
		{"hub rendered as docker.io", hubConfig(), []string{"docker.io"}},
		{"sorted across registries", `{"auths":{"quay.io":{},"ghcr.io":{},"https://index.docker.io/v1/":{}}}`, []string{"docker.io", "ghcr.io", "quay.io"}},
		{"legacy keys reduced to hostnames", `{"auths":{"https://ghcr.io/v1/":{}}}`, []string{"ghcr.io"}},
		{"duplicate hub keys collapsed", `{"auths":{"https://index.docker.io/v1/":{},"docker.io":{}}}`, []string{"docker.io"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := parseDockerConfig([]byte(tc.config))
			if err != nil {
				t.Fatalf("parseDockerConfig() errored: %s", err)
			}

			if got := authServersFromConfig(config); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("authServersFromConfig() = %v, want %v", got, tc.want)
			}
		})
	}
}
