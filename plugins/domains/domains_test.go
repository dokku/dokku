package domains

import (
	"io"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
	. "github.com/onsi/gomega"
)

// setupTestEnv points the dokku environment at temporary directories and
// creates the given apps
func setupTestEnv(t *testing.T, apps ...string) (dokkuRoot string, libRoot string) {
	t.Helper()

	libRoot = t.TempDir()
	dokkuRoot = t.TempDir()
	t.Setenv("DOKKU_LIB_ROOT", libRoot)
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	t.Setenv("PLUGIN_PATH", filepath.Join(libRoot, "plugins"))
	t.Setenv("PLUGIN_ENABLED_PATH", filepath.Join(libRoot, "plugins", "enabled"))
	t.Setenv("PLUGIN_CORE_AVAILABLE_PATH", filepath.Join(libRoot, "core-plugins", "available"))
	t.Setenv("DOKKU_QUIET_OUTPUT", "1")

	current, err := user.Current()
	Expect(err).To(Succeed())
	group, err := user.LookupGroupId(current.Gid)
	Expect(err).To(Succeed())
	t.Setenv("DOKKU_SYSTEM_USER", current.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", group.Name)

	Expect(common.PropertySetup("domains")).To(Succeed())
	for _, appName := range apps {
		Expect(os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755)).To(Succeed())
	}

	return dokkuRoot, libRoot
}

// writeVhost writes a VHOST file at the given path
func writeVhost(path string, contents string) {
	Expect(os.WriteFile(path, []byte(contents), 0644)).To(Succeed())
}

// captureStdout returns everything written to stdout while fn runs
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	reader, writer, err := os.Pipe()
	Expect(err).To(Succeed())

	original := os.Stdout
	os.Stdout = writer
	fnErr := fn()
	os.Stdout = original
	Expect(writer.Close()).To(Succeed())

	output, err := io.ReadAll(reader)
	Expect(err).To(Succeed())
	return string(output), fnErr
}

func TestReportAppEnabled(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t, "app-a")

	Expect(isAppVhostDisabled("app-a")).To(BeFalse())
	Expect(reportAppEnabled("app-a")).To(Equal("true"))

	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())
	Expect(isAppVhostDisabled("app-a")).To(BeTrue())
	Expect(reportAppEnabled("app-a")).To(Equal("false"))

	Expect(common.PropertyWrite("domains", "app-a", "disabled", "false")).To(Succeed())
	Expect(isAppVhostDisabled("app-a")).To(BeFalse())
	Expect(reportAppEnabled("app-a")).To(Equal("true"))
}

func TestReportAppVhosts(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, _ := setupTestEnv(t, "app-a")

	Expect(reportAppVhosts("app-a")).To(Equal(""))

	writeVhost(filepath.Join(dokkuRoot, "app-a", "VHOST"), "a.example.com\n\n  b.example.com\n")
	Expect(reportAppVhosts("app-a")).To(Equal("a.example.com b.example.com"))
}

func TestReportGlobal(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, _ := setupTestEnv(t)

	Expect(isGlobalVhostEnabled()).To(BeFalse())
	Expect(reportGlobalEnabled("")).To(Equal("false"))
	Expect(reportGlobalVhosts("")).To(Equal(""))

	writeVhost(filepath.Join(dokkuRoot, "VHOST"), "not a hostname!\n")
	Expect(isGlobalVhostEnabled()).To(BeFalse())
	Expect(reportGlobalVhosts("")).To(Equal(""))

	writeVhost(filepath.Join(dokkuRoot, "VHOST"), "dokku.me\ndokku.org\n")
	Expect(isGlobalVhostEnabled()).To(BeTrue())
	Expect(reportGlobalEnabled("")).To(Equal("true"))
	Expect(reportGlobalVhosts("")).To(Equal("dokku.me dokku.org"))
}

func TestReportSingleApp(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, _ := setupTestEnv(t, "app-a")
	writeVhost(filepath.Join(dokkuRoot, "VHOST"), "dokku.me\n")
	writeVhost(filepath.Join(dokkuRoot, "app-a", "VHOST"), "app-a.dokku.me\n")
	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())

	output, err := captureStdout(t, func() error {
		return ReportSingleApp("app-a", "", "--domains-app-enabled")
	})
	Expect(err).To(Succeed())
	Expect(output).To(Equal("false\n"))

	output, err = captureStdout(t, func() error {
		return ReportSingleApp("app-a", "", "--domains-app-vhosts")
	})
	Expect(err).To(Succeed())
	Expect(output).To(Equal("app-a.dokku.me\n"))

	output, err = captureStdout(t, func() error {
		return ReportSingleApp("app-a", "json", "")
	})
	Expect(err).To(Succeed())
	Expect(output).To(MatchJSON(`{"app-enabled":"false","app-vhosts":"app-a.dokku.me","global-enabled":"true","global-vhosts":"dokku.me"}`))

	output, err = captureStdout(t, func() error {
		return ReportSingleApp("--global", "json", "")
	})
	Expect(err).To(Succeed())
	Expect(output).To(MatchJSON(`{"global-enabled":"true","global-vhosts":"dokku.me"}`))

	_, err = captureStdout(t, func() error {
		return ReportSingleApp("--global", "", "--domains-app-enabled")
	})
	Expect(err).To(HaveOccurred())

	Expect(ReportSingleApp("missing-app", "", "")).To(HaveOccurred())
}

func TestCommandReport(t *testing.T) {
	RegisterTestingT(t)
	setupTestEnv(t)

	Expect(CommandReport("", "", "")).To(Succeed())
	Expect(CommandReport("missing-app", "", "")).To(HaveOccurred())

	setupTestEnv(t, "app-a", "app-b")
	Expect(common.PropertyWrite("domains", "app-b", "disabled", "true")).To(Succeed())

	output, err := captureStdout(t, func() error {
		return CommandReport("", "", "--domains-app-enabled")
	})
	Expect(err).To(Succeed())
	Expect(output).To(Equal("true\nfalse\n"))

	output, err = captureStdout(t, func() error {
		return CommandReport("app-b", "", "--domains-app-enabled")
	})
	Expect(err).To(Succeed())
	Expect(output).To(Equal("false\n"))

	_, err = captureStdout(t, func() error {
		return CommandReport("", "", "--invalid-flag")
	})
	Expect(err).To(HaveOccurred())
}

func TestHelpers(t *testing.T) {
	RegisterTestingT(t)

	Expect(readFileString(filepath.Join(t.TempDir(), "missing"))).To(Equal(""))
	Expect(normalizeWhitespace(" a\n\tb  c \n")).To(Equal("a b c"))
}
