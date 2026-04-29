package storage

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAttachmentValidate(t *testing.T) {
	RegisterTestingT(t)

	good := &Attachment{
		EntryName:     "demo-data",
		ContainerPath: "/data",
		Phases:        []string{PhaseDeploy, PhaseRun},
	}
	Expect(good.Validate()).To(Succeed())

	noEntry := &Attachment{ContainerPath: "/data", Phases: []string{PhaseDeploy}}
	Expect(noEntry.Validate()).To(HaveOccurred())

	noPath := &Attachment{EntryName: "demo-data", Phases: []string{PhaseDeploy}}
	Expect(noPath.Validate()).To(HaveOccurred())

	relative := &Attachment{EntryName: "demo-data", ContainerPath: "data", Phases: []string{PhaseDeploy}}
	Expect(relative.Validate()).To(HaveOccurred())

	noPhases := &Attachment{EntryName: "demo-data", ContainerPath: "/data"}
	Expect(noPhases.Validate()).To(HaveOccurred())

	badPhase := &Attachment{EntryName: "demo-data", ContainerPath: "/data", Phases: []string{"build"}}
	Expect(badPhase.Validate()).To(HaveOccurred())
}
