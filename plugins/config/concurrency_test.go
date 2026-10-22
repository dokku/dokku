package config

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
)

func TestConfigSetManyConcurrent(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	const count = 20
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = SetMany(testAppName, map[string]string{fmt.Sprintf("KEY%d", i): "value"}, false, false)
		}(i)
	}
	wg.Wait()
	Expect(errors.Join(errs...)).NotTo(HaveOccurred())

	for i := 0; i < count; i++ {
		expectValue(testAppName, fmt.Sprintf("KEY%d", i), "value")
	}
	expectValue(testAppName, "testKey", "TESTING")
}

func TestConfigUnsetManyConcurrent(t *testing.T) {
	RegisterTestingT(t)
	setupTestApp(t)

	const count = 20
	entries := map[string]string{}
	for i := 0; i < count; i++ {
		entries[fmt.Sprintf("KEY%d", i)] = "value"
	}
	Expect(SetMany(testAppName, entries, false, false)).To(Succeed())

	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = UnsetMany(testAppName, []string{fmt.Sprintf("KEY%d", i)}, false)
		}(i)
	}
	wg.Wait()
	Expect(errors.Join(errs...)).NotTo(HaveOccurred())

	for i := 0; i < count; i++ {
		expectNoValue(testAppName, fmt.Sprintf("KEY%d", i))
	}
	expectValue(testAppName, "testKey", "TESTING")
}
