package registry

import (
	"errors"
	"os/user"
	"sort"
	"sync"
	"testing"
)

func TestIncrementTagVersionConcurrent(t *testing.T) {
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	current, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current: %v", err)
	}
	group, err := user.LookupGroupId(current.Gid)
	if err != nil {
		t.Fatalf("user.LookupGroupId: %v", err)
	}
	t.Setenv("DOKKU_SYSTEM_USER", current.Username)
	t.Setenv("DOKKU_SYSTEM_GROUP", group.Name)

	const count = 20
	var wg sync.WaitGroup
	versions := make([]int, count)
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			versions[i], errs[i] = incrementTagVersion("race-test")
		}(i)
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("incrementTagVersion: %v", err)
	}

	sort.Ints(versions)
	for i, version := range versions {
		if version != i+1 {
			t.Fatalf("versions = %v, want 1 through %d with no duplicates", versions, count)
		}
	}
}
