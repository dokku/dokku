package traefikvhosts

import (
	"os/user"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

// setupPropertyEnv points the dokku env at temporary directories and tells the
// permission helpers to chown files to the current user (a no-op) so the test
// works without root.
func setupPropertyEnv(t *testing.T) {
	t.Helper()

	libRoot := t.TempDir()

	t.Setenv("DOKKU_LIB_ROOT", libRoot)
	t.Setenv("DOKKU_ROOT", t.TempDir())
	t.Setenv("PLUGIN_PATH", filepath.Join(libRoot, "plugins"))

	current, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current: %v", err)
	}
	group, err := user.LookupGroupId(current.Gid)
	if err != nil {
		t.Fatalf("user.LookupGroupId: %v", err)
	}
	t.Setenv("DOKKU_SYSTEM_USER", current.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", group.Name)
}

func staticReportFunc(value string) common.ReportFunc {
	return func(string) string {
		return value
	}
}

func TestMaskedReportFunc(t *testing.T) {
	const flagName = "--traefik-global-basic-auth-password"

	cases := []struct {
		name     string
		value    string
		format   string
		infoFlag string
		want     string
	}{
		{name: "stdout masks a set value", value: "hunter2", format: "stdout", want: valueMask},
		{name: "empty format masks a set value", value: "hunter2", want: valueMask},
		{name: "stdout leaves an unset value empty", value: "", format: "stdout", want: ""},
		{name: "json returns the raw value", value: "hunter2", format: "json", want: "hunter2"},
		{name: "matching info flag returns the raw value", value: "hunter2", format: "stdout", infoFlag: flagName, want: "hunter2"},
		{name: "other info flag masks the value", value: "hunter2", format: "stdout", infoFlag: "--traefik-global-basic-auth-username", want: valueMask},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fn := maskedReportFunc(staticReportFunc(tc.value), flagName, tc.format, tc.infoFlag)
			if got := fn("--global"); got != tc.want {
				t.Errorf("maskedReportFunc(%q, %q, %q) = %q, want %q", tc.value, tc.format, tc.infoFlag, got, tc.want)
			}
		})
	}
}

func TestDNSProviderReportFlags(t *testing.T) {
	setupPropertyEnv(t)

	if err := common.PropertyWrite("traefik", "--global", "dns-provider", "cloudflare"); err != nil {
		t.Fatalf("PropertyWrite dns-provider: %v", err)
	}
	if err := common.PropertyWrite("traefik", "--global", "dns-provider-cf_api_email", "test@example.com"); err != nil {
		t.Fatalf("PropertyWrite dns-provider-cf_api_email: %v", err)
	}
	if err := common.PropertyWrite("traefik", "--global", "dns-provider-cf_api_key", "secret-key"); err != nil {
		t.Fatalf("PropertyWrite dns-provider-cf_api_key: %v", err)
	}

	flags, err := dnsProviderReportFlags("stdout", "")
	if err != nil {
		t.Fatalf("dnsProviderReportFlags: %v", err)
	}

	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_email", valueMask)
	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_key", valueMask)
	if len(flags) != 2 {
		t.Errorf("flags = %v, want only the two dns-provider-* env var flags", flagNames(flags))
	}
	if _, ok := flags["--traefik-global-dns-provider"]; ok {
		t.Error("--traefik-global-dns-provider should be reported statically, not as a dynamic env var")
	}
	if _, ok := flags["--traefik-dns-provider-cf_api_email"]; ok {
		t.Error("--traefik-dns-provider-cf_api_email should not be emitted outside the global namespace")
	}

	flags, err = dnsProviderReportFlags("json", "")
	if err != nil {
		t.Fatalf("dnsProviderReportFlags: %v", err)
	}
	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_email", "test@example.com")
	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_key", "secret-key")

	flags, err = dnsProviderReportFlags("stdout", "--traefik-global-dns-provider-cf_api_key")
	if err != nil {
		t.Fatalf("dnsProviderReportFlags: %v", err)
	}
	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_email", valueMask)
	assertReportFuncValue(t, flags, "--traefik-global-dns-provider-cf_api_key", "secret-key")
}

func TestDNSProviderReportFlagsWithoutProperties(t *testing.T) {
	setupPropertyEnv(t)

	flags, err := dnsProviderReportFlags("stdout", "")
	if err != nil {
		t.Fatalf("dnsProviderReportFlags: %v", err)
	}
	if len(flags) != 0 {
		t.Errorf("flags = %v, want no flags", flagNames(flags))
	}
}

func assertReportFuncValue(t *testing.T, flags map[string]common.ReportFunc, flagName string, want string) {
	t.Helper()
	fn, ok := flags[flagName]
	if !ok {
		t.Fatalf("flag %q missing from %v", flagName, flagNames(flags))
	}
	if got := fn("--global"); got != want {
		t.Errorf("flag %q = %q, want %q", flagName, got, want)
	}
}

func flagNames(flags map[string]common.ReportFunc) []string {
	names := []string{}
	for name := range flags {
		names = append(names, name)
	}
	return names
}
