package storage

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestResolveChownIDRejectsOutOfBoundsValues(t *testing.T) {
	RegisterTestingT(t)

	for _, chownFlag := range []string{"-1", "65536", "100000", "231072", "abc", "1.5", ""} {
		_, err := ResolveChownID(chownFlag)
		Expect(err).To(HaveOccurred(), "expected %q to be rejected", chownFlag)
		Expect(err.Error()).To(ContainSubstring("Unsupported chown permissions"))
	}
}
