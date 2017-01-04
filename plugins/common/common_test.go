package common

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetEnv(t *testing.T) {
	RegisterTestingT(t)
	Expect(MustGetEnv("DOKKU_ROOT")).To(Equal("/home/dokku"))
}

func TestGetAppImageRepo(t *testing.T) {
	RegisterTestingT(t)
	Expect(GetAppImageRepo("testapp")).To(Equal("dokku/testapp"))
}

func TestVerifyImageInvalid(t *testing.T) {
	RegisterTestingT(t)
	Expect(VerifyImage("testapp")).To(Equal(false))
}

func TestVerifyAppNameInvalid(t *testing.T) {
	RegisterTestingT(t)
	err := VerifyAppName("1994testApp")
	Expect(err).To(HaveOccurred())
}

func TestVerifyAppName(t *testing.T) {
	RegisterTestingT(t)
	dir := "/home/dokku/testApp"
	os.MkdirAll(dir, 0644)
	err := VerifyAppName("testApp")
	Expect(err).NotTo(HaveOccurred())
	os.RemoveAll(dir)
}
