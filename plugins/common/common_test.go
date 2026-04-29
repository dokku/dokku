package common

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

var (
	testAppName  = "test-app-1"
	testAppDir   = strings.Join([]string{"/home/dokku/", testAppName}, "")
	testEnvFile  = strings.Join([]string{testAppDir, "/ENV"}, "")
	testEnvLine  = "export testKey=TESTING"
	testAppName2 = "01-test-app-1"
	testAppDir2  = strings.Join([]string{"/home/dokku/", testAppName2}, "")
	testEnvFile2 = strings.Join([]string{testAppDir2, "/ENV"}, "")
	testEnvLine2 = "export testKey=TESTING"
)

func setupTests() (err error) {
	if err := os.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins"); err != nil {
		return err
	}

	return os.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")
}

func setupTestApp() (err error) {
	Expect(os.MkdirAll(testAppDir, 0766)).To(Succeed())
	b := []byte(testEnvLine + "\n")
	if err = os.WriteFile(testEnvFile, b, 0644); err != nil {
		return
	}
	return
}

func setupTestApp2() (err error) {
	Expect(os.MkdirAll(testAppDir2, 0766)).To(Succeed())
	b := []byte(testEnvLine2 + "\n")
	if err = os.WriteFile(testEnvFile2, b, 0644); err != nil {
		return
	}
	return
}

func teardownTestApp() {
	os.RemoveAll(testAppDir)
}

func teardownTestApp2() {
	os.RemoveAll(testAppDir2)
}

func TestCommonGetEnv(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(MustGetEnv("DOKKU_ROOT")).To(Equal("/home/dokku"))
}

func TestCommonGetAppImageRepo(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(GetAppImageRepo("testapp")).To(Equal("dokku/testapp"))
}

func TestCommonVerifyImageInvalid(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(VerifyImage("testapp")).To(Equal(false))
}

func TestCommonVerifyAppNameInvalid(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(VerifyAppName("1994testApp")).To(HaveOccurred())
}

func TestCommonVerifyAppName(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(setupTestApp()).To(Succeed())
	Expect(VerifyAppName(testAppName)).To(Succeed())
	teardownTestApp()

	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(setupTestApp2()).To(Succeed())
	Expect(VerifyAppName(testAppName2)).To(Succeed())
	teardownTestApp2()
}

func TestCommonDokkuAppsError(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	_, err := DokkuApps()
	Expect(err).To(HaveOccurred())
}

func TestCommonDokkuApps(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	Expect(setupTestApp()).To(Succeed())
	apps, err := DokkuApps()
	Expect(err).NotTo(HaveOccurred())
	Expect(apps).To(HaveLen(1))
	Expect(apps[0]).To(Equal(testAppName))
	teardownTestApp()
}

func TestCommonStripInlineComments(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTests()).To(Succeed())
	text := StripInlineComments(strings.Join([]string{testEnvLine, "# testing comment"}, " "))
	Expect(text).To(Equal(testEnvLine))
}

func TestParseReportArgsEmpty(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeFalse())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(BeEmpty())
}

func TestParseReportArgsAppOnly(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"myapp"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeFalse())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(Equal([]string{"myapp"}))
}

func TestParseReportArgsFormatJSON(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"myapp", "--format", "json"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeFalse())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(Equal([]string{"myapp", "--format", "json"}))
}

func TestParseReportArgsGlobalAlone(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"--global"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeTrue())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(BeEmpty())
}

func TestParseReportArgsGlobalWithFormat(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"--global", "--format", "json"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeTrue())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(Equal([]string{"--format", "json"}))
}

func TestParseReportArgsInfoFlag(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"myapp", "--scheduler-selected"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeFalse())
	Expect(args.InfoFlag).To(Equal("--scheduler-selected"))
	Expect(args.OSArgs).To(Equal([]string{"myapp"}))
}

func TestParseReportArgsAppThenGlobal(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"myapp", "--global"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeTrue())
	Expect(args.InfoFlag).To(Equal(""))
	Expect(args.OSArgs).To(Equal([]string{"myapp"}))
}

func TestParseReportArgsGlobalWithInfoFlag(t *testing.T) {
	RegisterTestingT(t)
	args, err := ParseReportArgs("scheduler", []string{"--global", "--scheduler-global-selected"})
	Expect(err).NotTo(HaveOccurred())
	Expect(args.IsGlobal).To(BeTrue())
	Expect(args.InfoFlag).To(Equal("--scheduler-global-selected"))
	Expect(args.OSArgs).To(BeEmpty())
}

func TestParseReportArgsMultipleInfoFlags(t *testing.T) {
	RegisterTestingT(t)
	_, err := ParseReportArgs("scheduler", []string{"--scheduler-selected", "--scheduler-global-selected"})
	Expect(err).To(HaveOccurred())
}
