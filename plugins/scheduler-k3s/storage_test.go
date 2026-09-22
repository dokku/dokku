package scheduler_k3s

import (
	"testing"

	"github.com/dokku/dokku/plugins/storage"
)

func TestValidateMountPairsRejectsDockerLocalEntry(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry: &storage.Entry{
				Name:      "demo-data",
				Scheduler: storage.SchedulerDockerLocal,
			},
			Attachment: &storage.Attachment{
				ContainerPath: "/data",
			},
		},
	}

	err := ValidateMountPairs(pairs)
	if err == nil {
		t.Fatal("expected error for docker-local entry, got nil")
	}
	expected := `storage entry "demo-data" is scheduler=docker-local but is mounted on a k3s app; recreate it with --scheduler k3s`
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateMountPairsRejectsNonK3sEntry(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry: &storage.Entry{
				Name:      "demo-custom",
				Scheduler: "nomad",
			},
			Attachment: &storage.Attachment{
				ContainerPath: "/data",
			},
		},
	}

	err := ValidateMountPairs(pairs)
	if err == nil {
		t.Fatal("expected error for non-k3s entry, got nil")
	}
	expected := `storage entry "demo-custom" is scheduler=nomad but is mounted on a k3s app; recreate it with --scheduler k3s`
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

// TestValidateMountPairsRejectsOutOfScopeEntry pins the reason validation is a
// separate pass: a mismatched entry scoped to a process type the caller is not
// currently building still has to fail the deploy.
func TestValidateMountPairsRejectsOutOfScopeEntry(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry: &storage.Entry{
				Name:      "demo-data",
				Scheduler: storage.SchedulerDockerLocal,
			},
			Attachment: &storage.Attachment{
				ContainerPath: "/data",
				ProcessType:   "worker",
			},
		},
	}

	if err := ValidateMountPairs(pairs); err == nil {
		t.Fatal("expected error for docker-local entry scoped to another process, got nil")
	}
}

func TestToProcessVolumesAcceptsK3sEntry(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry: &storage.Entry{
				Name:      "demo-pvc",
				Scheduler: storage.SchedulerK3s,
			},
			Attachment: &storage.Attachment{
				ContainerPath: "/app/data",
				Subpath:       "sub",
				Readonly:      true,
			},
		},
	}

	volumes := ToProcessVolumes(pairs, "web")
	if len(volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(volumes))
	}
	v := volumes[0]
	if v.Name != "demo-pvc" {
		t.Errorf("expected volume name demo-pvc, got %s", v.Name)
	}
	if v.MountPath != "/app/data" {
		t.Errorf("expected mount path /app/data, got %s", v.MountPath)
	}
	if v.SubPath != "sub" {
		t.Errorf("expected subpath sub, got %s", v.SubPath)
	}
	if !v.ReadOnly {
		t.Errorf("expected read only true, got %v", v.ReadOnly)
	}
	if v.PersistentClaim == nil || v.PersistentClaim.ClaimName != "demo-pvc" {
		t.Errorf("expected persistent claim name demo-pvc, got %+v", v.PersistentClaim)
	}
}

// TestToProcessVolumesCollapsesOneEntryMountedTwice covers the pod-volume
// uniqueness rule: mounting one storage entry at two container paths has to
// produce two volumeMounts and a single volume, or the API server rejects the
// chart with a duplicate volume name.
func TestToProcessVolumesCollapsesOneEntryMountedTwice(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry:      &storage.Entry{Name: "demo-pvc", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/one", Subpath: "one"},
		},
		{
			Entry:      &storage.Entry{Name: "demo-pvc", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/two", Subpath: "two", Readonly: true},
		},
	}

	volumes := ToProcessVolumes(pairs, "web")
	if len(volumes) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(volumes))
	}
	if volumes[0].Name != volumes[1].Name {
		t.Fatalf("expected both entries to name one volume, got %q and %q", volumes[0].Name, volumes[1].Name)
	}
	if volumes[0].MountPath != "/one" || volumes[1].MountPath != "/two" {
		t.Fatalf("expected mounts at /one and /two, got %q and %q", volumes[0].MountPath, volumes[1].MountPath)
	}
	if !volumes[1].ReadOnly || volumes[0].ReadOnly {
		t.Errorf("expected only the second mount to be read only")
	}
}

// TestProcessVolumesForRejectsSameNameDifferentSource guards the name collapse
// the chart performs: the pod volume is emitted from the first entry with a
// given name, so two different sources sharing one name would silently mount
// whichever came first. An app with a storage entry called "shmem" alongside a
// configured shm size is the way to reach it.
func TestProcessVolumesForRejectsSameNameDifferentSource(t *testing.T) {
	base := []ProcessVolume{
		{
			Name:      "shmem",
			MountPath: "/dev/shm",
			EmptyDir:  &ProcessVolumeEmptyDir{Medium: "Memory", SizeLimit: "64Mi"},
		},
	}

	pairs := []AppMountPair{
		{
			Entry:      &storage.Entry{Name: "shmem", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/data"},
		},
	}

	if _, err := processVolumesFor(base, pairs, "web"); err == nil {
		t.Fatal("expected an error for two sources claiming one volume name, got nil")
	}
}

func TestToProcessVolumesSkipsNilPairs(t *testing.T) {
	pairs := []AppMountPair{
		{Entry: nil, Attachment: nil},
		{Entry: &storage.Entry{Name: "demo", Scheduler: storage.SchedulerK3s}, Attachment: nil},
		{Entry: nil, Attachment: &storage.Attachment{ContainerPath: "/data"}},
	}

	if volumes := ToProcessVolumes(pairs, "web"); len(volumes) != 0 {
		t.Fatalf("expected 0 volumes, got %d", len(volumes))
	}
}

// TestToProcessVolumesScopesToProcessType is the k3s half of #9059: a mount
// scoped to a named process type reaches that process and no other, while a
// default-scoped mount reaches all of them.
func TestToProcessVolumesScopesToProcessType(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry:      &storage.Entry{Name: "shared", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/shared", ProcessType: storage.DefaultProcessType},
		},
		{
			Entry:      &storage.Entry{Name: "web-only", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/web", ProcessType: "web"},
		},
		{
			Entry:      &storage.Entry{Name: "legacy", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/legacy"},
		},
	}

	for _, tc := range []struct {
		processType string
		expected    []string
	}{
		{processType: "web", expected: []string{"shared", "web-only", "legacy"}},
		{processType: "worker", expected: []string{"shared", "legacy"}},
		{processType: "", expected: []string{"shared", "legacy"}},
	} {
		names := []string{}
		for _, volume := range ToProcessVolumes(pairs, tc.processType) {
			names = append(names, volume.Name)
		}
		if len(names) != len(tc.expected) {
			t.Fatalf("process type %q: expected %v, got %v", tc.processType, tc.expected, names)
		}
		for i, name := range tc.expected {
			if names[i] != name {
				t.Fatalf("process type %q: expected %v, got %v", tc.processType, tc.expected, names)
			}
		}
	}
}

// TestProcessVolumesForDoesNotAliasBase pins that each process type gets its
// own slice: appending onto a shared backing array would leak one process's
// storage mounts into another's pod spec.
func TestProcessVolumesForDoesNotAliasBase(t *testing.T) {
	base := make([]ProcessVolume, 0, 8)
	base = append(base, ProcessVolume{Name: "shmem", MountPath: "/dev/shm", EmptyDir: &ProcessVolumeEmptyDir{Medium: "Memory"}})

	pairs := []AppMountPair{
		{
			Entry:      &storage.Entry{Name: "web-only", Scheduler: storage.SchedulerK3s},
			Attachment: &storage.Attachment{ContainerPath: "/web", ProcessType: "web"},
		},
	}

	web, err := processVolumesFor(base, pairs, "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	worker, err := processVolumesFor(base, pairs, "worker")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(web) != 2 {
		t.Fatalf("expected web to get 2 volumes, got %d", len(web))
	}
	if len(worker) != 1 {
		t.Fatalf("expected worker to get 1 volume, got %d", len(worker))
	}
	if worker[0].Name != "shmem" {
		t.Fatalf("expected worker to keep only the base volume, got %s", worker[0].Name)
	}
	if len(base) != 1 {
		t.Fatalf("expected base to be left untouched, got %d volumes", len(base))
	}
}
