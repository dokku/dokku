package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// exit codes reported by registry:auth-status. Codes 0 through 3 match the
// codes git:auth-status reports for the equivalent netrc states.
const (
	// authStatusMatch means the stored credential matches the requested state
	authStatusMatch = 0

	// authStatusMissing means no credential is stored for the server
	authStatusMissing = 1

	// authStatusDiffers means a credential is stored but does not match
	authStatusDiffers = 2

	// authStatusInvalidArguments means the state could not be checked
	authStatusInvalidArguments = 3

	// authStatusUnknown means a credential is stored but its secret is unreadable
	authStatusUnknown = 4
)

// dockerAuthEntry is the subset of a docker config.json auths entry dokku reads
type dockerAuthEntry struct {
	Auth          string `json:"auth"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	IdentityToken string `json:"identitytoken"`
}

// dockerConfigFile is the subset of a docker config.json dokku reads
type dockerConfigFile struct {
	Auths       map[string]dockerAuthEntry `json:"auths"`
	CredsStore  string                     `json:"credsStore"`
	CredHelpers map[string]string          `json:"credHelpers"`
}

// authStatusInput describes the credential state a caller is asking about
type authStatusInput struct {
	// ConfigBytes is the contents of the docker config.json, nil when absent
	ConfigBytes []byte

	// Server is the registry server being asked about
	Server string

	// Username is the expected username, empty to assert no credential exists
	Username string

	// Password is the expected password
	Password string
}

// parseDockerConfig unmarshals the subset of a docker config.json dokku reads
func parseDockerConfig(b []byte) (dockerConfigFile, error) {
	var config dockerConfigFile
	if len(b) == 0 {
		return config, nil
	}

	if err := json.Unmarshal(b, &config); err != nil {
		return config, err
	}

	return config, nil
}

// lookupAuthEntry returns the auths entry docker would use for a server. The
// exact key is tried first, then every stored key is compared in hostname form,
// which is how docker matches credentials written under a legacy key.
func lookupAuthEntry(config dockerConfigFile, server string) (dockerAuthEntry, bool) {
	key := dockerAuthKey(server)
	if entry, ok := config.Auths[key]; ok {
		return entry, true
	}

	// map iteration order is random, so the keys are sorted to keep the answer
	// stable when more than one of them resolves to the same server
	storedKeys := make([]string, 0, len(config.Auths))
	for storedKey := range config.Auths {
		storedKeys = append(storedKeys, storedKey)
	}
	sort.Strings(storedKeys)

	for _, storedKey := range storedKeys {
		if convertToHostname(storedKey) == key {
			return config.Auths[storedKey], true
		}
	}

	return dockerAuthEntry{}, false
}

// credentialHelperFor returns the credential helper docker would consult for a
// server, preferring a server-specific helper over the global store
func credentialHelperFor(config dockerConfigFile, server string) string {
	key := dockerAuthKey(server)
	if helper, ok := config.CredHelpers[key]; ok && helper != "" {
		return helper
	}

	hostname := convertToHostname(normalizeRegistryServer(server))
	if helper, ok := config.CredHelpers[hostname]; ok && helper != "" {
		return helper
	}

	return config.CredsStore
}

// decodeAuthEntry returns the username and password stored in an auths entry.
// The secret is unreadable when a helper holds it or the registry issued an
// identity token, in which case a username may still be available.
func decodeAuthEntry(entry dockerAuthEntry) (string, string, bool) {
	if entry.Auth == "" {
		if entry.Username != "" && entry.Password != "" {
			return entry.Username, entry.Password, true
		}

		return entry.Username, "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(entry.Auth)
	if err != nil {
		return "", "", false
	}

	username, password, found := strings.Cut(string(decoded), ":")
	if !found {
		// without a separator there is no username to read either
		return "", "", false
	}

	if password == "" {
		return username, "", false
	}

	return username, password, true
}

// checkAuthStatus reports how a stored docker credential compares to the
// requested state. It returns the exit code to terminate with and, when the
// credential cannot be compared, the reason to print to stderr.
func checkAuthStatus(input authStatusInput) (int, string) {
	config, err := parseDockerConfig(input.ConfigBytes)
	if err != nil {
		return authStatusUnknown, fmt.Sprintf("Unable to parse the docker config for %s: %s", input.Server, err.Error())
	}

	entry, ok := lookupAuthEntry(config, input.Server)
	if !ok {
		if input.Username == "" {
			return authStatusMatch, ""
		}

		return authStatusMissing, ""
	}

	// that a credential exists is the whole answer an empty username asks for,
	// and needs no access to the secret it holds
	if input.Username == "" {
		return authStatusDiffers, ""
	}

	// a helper keeps the secret outside of the file, leaving nothing to compare
	// against in the entry docker blanked when it stored the credential
	if helper := credentialHelperFor(config, input.Server); helper != "" {
		return authStatusUnknown, fmt.Sprintf("Unable to compare the credential for %s: the %s credential helper holds it", input.Server, helper)
	}

	username, password, readable := decodeAuthEntry(entry)
	if readable {
		if username != input.Username || password != input.Password {
			return authStatusDiffers, ""
		}

		return authStatusMatch, ""
	}

	// a username that differs settles the comparison without the secret, so
	// only an otherwise matching credential is left indeterminate
	if username != "" && username != input.Username {
		return authStatusDiffers, ""
	}

	if entry.IdentityToken != "" {
		return authStatusUnknown, fmt.Sprintf("Unable to compare the credential for %s: it is stored as an identity token", input.Server)
	}

	return authStatusUnknown, fmt.Sprintf("Unable to compare the credential for %s: no password is stored for it", input.Server)
}

// authServersFromConfig returns the sorted registry servers a docker config
// holds credentials for, named as they would be passed to registry:login
func authServersFromConfig(config dockerConfigFile) []string {
	servers := []string{}
	seen := map[string]bool{}
	for key := range config.Auths {
		server := convertToHostname(key)
		if dockerAuthKey(key) == dockerIndexServer {
			server = "docker.io"
		}

		if server == "" || seen[server] {
			continue
		}

		seen[server] = true
		servers = append(servers, server)
	}

	sort.Strings(servers)
	return servers
}
