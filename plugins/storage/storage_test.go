package storage

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestVerifyPathsAbsolutePath(t *testing.T) {
	RegisterTestingT(t)

	Expect(VerifyPaths("/host/path:/container/path")).To(Succeed())
	Expect(VerifyPaths("/var/lib/dokku/data/storage/test:/app/data")).To(Succeed())
}

func TestVerifyPathsNamedVolume(t *testing.T) {
	RegisterTestingT(t)

	Expect(VerifyPaths("volume_name:/container/path")).To(Succeed())
	Expect(VerifyPaths("my-volume:/app/data")).To(Succeed())
	Expect(VerifyPaths("my.volume:/app/data")).To(Succeed())
}

func TestVerifyPathsInvalid(t *testing.T) {
	RegisterTestingT(t)

	err := VerifyPaths("/host/path")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Storage path must be two valid paths divided by colon"))

	err = VerifyPaths("a:/container")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Volume name must be two characters or more"))

	err = VerifyPaths("-invalid:/container")
	Expect(err).To(HaveOccurred())
}

func TestValidateDirectoryName(t *testing.T) {
	RegisterTestingT(t)

	Expect(ValidateDirectoryName("myapp")).To(Succeed())
	Expect(ValidateDirectoryName("my-app")).To(Succeed())
	Expect(ValidateDirectoryName("my_app")).To(Succeed())
	Expect(ValidateDirectoryName("MyApp123")).To(Succeed())
}

func TestValidateDirectoryNameInvalid(t *testing.T) {
	RegisterTestingT(t)

	err := ValidateDirectoryName("")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Please specify a directory"))

	err = ValidateDirectoryName("@invalid")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Directory can only contain"))

	err = ValidateDirectoryName("my/app")
	Expect(err).To(HaveOccurred())

	err = ValidateDirectoryName("my app")
	Expect(err).To(HaveOccurred())
}

func TestParseMountPath(t *testing.T) {
	RegisterTestingT(t)

	entry := ParseMountPath("/host/path:/container/path")
	Expect(entry.HostPath).To(Equal("/host/path"))
	Expect(entry.ContainerPath).To(Equal("/container/path"))
	Expect(entry.VolumeOptions).To(BeEmpty())

	entry = ParseMountPath("/host/path:/container/path:ro")
	Expect(entry.HostPath).To(Equal("/host/path"))
	Expect(entry.ContainerPath).To(Equal("/container/path"))
	Expect(entry.VolumeOptions).To(Equal("ro"))

	entry = ParseMountPath("volume_name:/container/path")
	Expect(entry.HostPath).To(Equal("volume_name"))
	Expect(entry.ContainerPath).To(Equal("/container/path"))
	Expect(entry.VolumeOptions).To(BeEmpty())
}

func TestGetStorageDirectory(t *testing.T) {
	RegisterTestingT(t)

	dir := GetStorageDirectory()
	Expect(dir).To(ContainSubstring("data/storage"))
}
