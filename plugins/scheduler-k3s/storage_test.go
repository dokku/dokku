package scheduler_k3s

import (
	"testing"

	"github.com/dokku/dokku/plugins/storage"
)

func TestToProcessVolumesRejectsDockerLocalEntry(t *testing.T) {
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

	_, err := ToProcessVolumes(pairs)
	if err == nil {
		t.Fatal("expected error for docker-local entry, got nil")
	}
	expected := `storage entry "demo-data" is scheduler=docker-local but is mounted on a k3s app; recreate it with --scheduler k3s`
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestToProcessVolumesRejectsNonK3sEntry(t *testing.T) {
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

	_, err := ToProcessVolumes(pairs)
	if err == nil {
		t.Fatal("expected error for non-k3s entry, got nil")
	}
	expected := `storage entry "demo-custom" is scheduler=nomad but is mounted on a k3s app; recreate it with --scheduler k3s`
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
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

	volumes, err := ToProcessVolumes(pairs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestToProcessVolumesSkipsNilPairs(t *testing.T) {
	pairs := []AppMountPair{
		{Entry: nil, Attachment: nil},
		{Entry: &storage.Entry{Name: "demo", Scheduler: storage.SchedulerK3s}, Attachment: nil},
		{Entry: nil, Attachment: &storage.Attachment{ContainerPath: "/data"}},
	}

	volumes, err := ToProcessVolumes(pairs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(volumes) != 0 {
		t.Fatalf("expected 0 volumes, got %d", len(volumes))
	}
}
