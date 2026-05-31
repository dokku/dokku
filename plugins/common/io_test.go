package common

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCommonFileToSlice(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	lines, err := FileToSlice(testEnvFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(lines).To(Equal([]string{testEnvLine}))
	teardownTestApp()
}

func TestCommonFileExists(t *testing.T) {
	RegisterTestingT(t)
	Expect(setupTestApp()).To(Succeed())
	Expect(FileExists(testEnvFile)).To(BeTrue())
	teardownTestApp()
}

func TestCommonReadFirstLine(t *testing.T) {
	RegisterTestingT(t)
	line := ReadFirstLine(testEnvFile)
	Expect(line).To(Equal(""))
	Expect(setupTestApp()).To(Succeed())
	line = ReadFirstLine(testEnvFile)
	Expect(line).To(Equal(testEnvLine))
	teardownTestApp()
}

// setupTouchFileTest pins DOKKU_SYSTEM_USER/GROUP to the running test user so
// SetPermissions' chown succeeds for non-root test runs.
func setupTouchFileTest(t *testing.T) {
	t.Helper()
	u, err := user.Current()
	Expect(err).NotTo(HaveOccurred())
	g, err := user.LookupGroupId(u.Gid)
	Expect(err).NotTo(HaveOccurred())
	t.Setenv("DOKKU_SYSTEM_USER", u.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", g.Name)
}

func TestCommonTouchFileCreatesMissingFile(t *testing.T) {
	RegisterTestingT(t)
	setupTouchFileTest(t)

	path := filepath.Join(t.TempDir(), "new-file")
	Expect(FileExists(path)).To(BeFalse())
	Expect(TouchFile(path)).To(Succeed())

	info, err := os.Stat(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(info.Size()).To(BeNumerically("==", 0))
	Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
}

// TestCommonTouchFilePreservesExistingContents is a regression test for #8722:
// scheduler-k3s:cluster:add used to zero out the dokku user's ~/.ssh/known_hosts
// because TouchFile opened the file with O_TRUNC.
func TestCommonTouchFilePreservesExistingContents(t *testing.T) {
	RegisterTestingT(t)
	setupTouchFileTest(t)

	path := filepath.Join(t.TempDir(), "existing-file")
	contents := []byte("host ssh-rsa AAAA...\n")
	Expect(os.WriteFile(path, contents, 0600)).To(Succeed())

	Expect(TouchFile(path)).To(Succeed())

	got, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(contents))
}
