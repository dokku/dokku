package common

import (
	"io/ioutil"
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
	Expect(os.MkdirAll(testAppDir, 0644)).To(Succeed())
	b := []byte(testEnvLine + "\n")
	if err = ioutil.WriteFile(testEnvFile, b, 0644); err != nil {
		return
	}
	return
}

func setupTestApp2() (err error) {
	Expect(os.MkdirAll(testAppDir2, 0644)).To(Succeed())
	b := []byte(testEnvLine2 + "\n")
	if err = ioutil.WriteFile(testEnvFile2, b, 0644); err != nil {
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
