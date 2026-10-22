package storage

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
)

func TestAddAttachmentConcurrent(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	const count = 20
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = AddAttachment("demo", &Attachment{
				EntryName:     fmt.Sprintf("demo-data-%d", i),
				ContainerPath: fmt.Sprintf("/data/%d", i),
				Phases:        []string{PhaseDeploy, PhaseRun},
				ProcessType:   DefaultProcessType,
			})
		}(i)
	}
	wg.Wait()
	Expect(errors.Join(errs...)).NotTo(HaveOccurred())

	attachments, err := LoadAttachments("demo")
	Expect(err).NotTo(HaveOccurred())
	Expect(attachments).To(HaveLen(count))
}

func TestUpdateEntryConcurrent(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-data",
		Scheduler: SchedulerK3s,
		Size:      "2Gi",
	})).To(Succeed())

	const count = 20
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = UpdateEntry("demo-data", func(entry *Entry) error {
				if entry.Labels == nil {
					entry.Labels = map[string]string{}
				}
				entry.Labels[fmt.Sprintf("label-%d", i)] = "value"
				return nil
			})
		}(i)
	}
	wg.Wait()
	Expect(errors.Join(errs...)).NotTo(HaveOccurred())

	entry, err := LoadEntry("demo-data")
	Expect(err).NotTo(HaveOccurred())
	Expect(entry.Labels).To(HaveLen(count))
}

func TestUpdateEntryErrorLeavesEntryUnchanged(t *testing.T) {
	RegisterTestingT(t)
	withTempLibRoot(t)

	Expect(SaveEntry(&Entry{
		Name:      "demo-data",
		Scheduler: SchedulerK3s,
		Size:      "2Gi",
	})).To(Succeed())

	_, err := UpdateEntry("demo-data", func(entry *Entry) error {
		entry.Size = "5Gi"
		return errors.New("boom")
	})
	Expect(err).To(MatchError("boom"))

	entry, err := LoadEntry("demo-data")
	Expect(err).NotTo(HaveOccurred())
	Expect(entry.Size).To(Equal("2Gi"))
}
