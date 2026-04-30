package scheduler_k3s

import (
	"testing"

	"github.com/dokku/dokku/plugins/storage"
)

func TestBuildStoragePodSpec(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerK3s,
		Namespace: "dokku",
	}
	pod, err := buildStoragePodSpec("dokku-storage-exec-demo-data-1", "dokku", entry, StorageExecInput{
		Image:       "alpine:3",
		Interactive: true,
		Tty:         true,
		Command:     []string{"sh"},
	})
	if err != nil {
		t.Fatalf("buildStoragePodSpec err: %v", err)
	}

	if pod.Spec.RestartPolicy != "Never" {
		t.Fatalf("expected RestartPolicy=Never, got %s", pod.Spec.RestartPolicy)
	}
	if len(pod.Spec.Volumes) != 1 || pod.Spec.Volumes[0].PersistentVolumeClaim == nil {
		t.Fatalf("expected one PVC volume, got %+v", pod.Spec.Volumes)
	}
	if pod.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "demo-data" {
		t.Fatalf("expected PVC claim name demo-data, got %s", pod.Spec.Volumes[0].PersistentVolumeClaim.ClaimName)
	}
	if len(pod.Spec.Containers) != 1 {
		t.Fatalf("expected one container, got %d", len(pod.Spec.Containers))
	}
	c := pod.Spec.Containers[0]
	if c.Image != "alpine:3" {
		t.Fatalf("expected image alpine:3, got %s", c.Image)
	}
	if !c.Stdin || !c.TTY {
		t.Fatalf("expected stdin and tty true on the container, got stdin=%v tty=%v", c.Stdin, c.TTY)
	}
	if len(c.VolumeMounts) != 1 || c.VolumeMounts[0].MountPath != "/data" {
		t.Fatalf("expected one volume mount at /data, got %+v", c.VolumeMounts)
	}
}

func TestBuildStoragePodSpecChownToSecurityContext(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerK3s,
		Chown:     "herokuish",
	}
	pod, err := buildStoragePodSpec("p", "default", entry, StorageExecInput{Image: "alpine:3"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsUser == nil {
		t.Fatalf("expected SecurityContext.RunAsUser to be populated for chown=herokuish")
	}
	// Don't assert exact UID because ResolveChownID consults docker for
	// userns offset; only assert non-nil and non-zero.
	if *pod.Spec.SecurityContext.RunAsUser == 0 {
		t.Fatalf("expected non-zero UID for chown=herokuish, got 0")
	}
}

func TestBuildStoragePodSpecAsUserOverride(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerK3s,
		Chown:     "herokuish",
	}
	pod, err := buildStoragePodSpec("p", "default", entry, StorageExecInput{
		Image:  "alpine:3",
		AsUser: "0",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pod.Spec.SecurityContext == nil || pod.Spec.SecurityContext.RunAsUser == nil {
		t.Fatalf("expected RunAsUser to be set")
	}
	if *pod.Spec.SecurityContext.RunAsUser != 0 {
		t.Fatalf("expected RunAsUser to be 0 (override), got %d", *pod.Spec.SecurityContext.RunAsUser)
	}
}

func TestBuildStoragePodSpecChownFalseLeavesDefaultUser(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerK3s,
		Chown:     "false",
	}
	pod, err := buildStoragePodSpec("p", "default", entry, StorageExecInput{Image: "alpine:3"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if pod.Spec.SecurityContext != nil {
		t.Fatalf("expected no SecurityContext when chown=false, got %+v", pod.Spec.SecurityContext)
	}
}
