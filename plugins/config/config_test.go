package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

var (
	testAppName = "test-app-1"
	testAppDir  = strings.Join([]string{"/home/dokku/", testAppName}, "")
)

func setupTestApp() (err error) {
	Expect(os.MkdirAll(testAppDir, 0644)).To(Succeed())
	b := []byte("export testKey=TESTING\n")
	if err = ioutil.WriteFile(strings.Join([]string{testAppDir, "/ENV"}, ""), b, 0644); err != nil {
		return
	}
	return
}

func teardownTestApp() {
	os.RemoveAll(testAppDir)
}

func TestConfigGetWithDefault(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	Expect(GetWithDefault(testAppName, "unknownKey", "UNKNOWN")).To(Equal("UNKNOWN"))
	Expect(GetWithDefault(testAppName, "testKey", "testKey")).To(Equal("TESTING"))
	teardownTestApp()
}
