package common

import (
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
