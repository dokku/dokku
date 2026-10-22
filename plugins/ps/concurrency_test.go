package ps

import (
	"errors"
	"fmt"
	"os/user"
	"sync"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

func TestUpdateScaleConcurrent(t *testing.T) {
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

	const appName = "race-test"
	const count = 20
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = updateScale(appName, false, FormationSlice{
				{ProcessType: fmt.Sprintf("worker%d", i), Quantity: 1},
			})
		}(i)
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("updateScale: %v", err)
	}

	tuples, err := common.PropertyListGet("ps", appName, "scale")
	if err != nil {
		t.Fatalf("PropertyListGet: %v", err)
	}
	if len(tuples) != count {
		t.Fatalf("scale has %d process types, want %d: %v", len(tuples), count, tuples)
	}
}
