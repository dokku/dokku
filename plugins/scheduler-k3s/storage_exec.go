package scheduler_k3s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/storage"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilexec "k8s.io/client-go/util/exec"
)

// StorageExecInput captures the inputs forwarded by the storage plugin
// over the scheduler-storage-exec trigger.
type StorageExecInput struct {
	EntryName   string
	Image       string
	Interactive bool
	Tty         bool
	AsUser      string
	Command     []string
}

// TriggerSchedulerStorageExec runs an interactive or non-interactive
// command against a k3s storage entry by creating a throwaway Pod that
// mounts the PVC at /data, exec-ing the user's command via the
// kubernetes API, and deleting the Pod afterwards. Errors at every
// stage (PVC missing, image pull failure, etc.) come through the
// structured Kubernetes API objects rather than kubectl-shaped strings.
func TriggerSchedulerStorageExec(ctx context.Context, scheduler string, input StorageExecInput) error {
	if scheduler != storage.SchedulerK3s {
		return nil
	}
	if !storage.EntryExists(input.EntryName) {
		return fmt.Errorf("storage entry %q does not exist", input.EntryName)
	}
	entry, err := storage.LoadEntry(input.EntryName)
	if err != nil {
		return err
	}
	if entry.Scheduler != storage.SchedulerK3s {
		return fmt.Errorf("storage entry %q has scheduler %q, not k3s", entry.Name, entry.Scheduler)
	}

	if err := isKubernetesAvailable(); err != nil {
		return fmt.Errorf("kubernetes not available: %w", err)
	}

	clientset, err := NewKubernetesClient()
	if err != nil {
		return err
	}

	namespace := entry.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if _, err := clientset.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, entry.Name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("PVC %s not found in namespace %s", entry.Name, namespace)
		}
		return fmt.Errorf("error reading PVC %s in namespace %s: %w", entry.Name, namespace, err)
	}

	podName := fmt.Sprintf("dokku-storage-exec-%s-%d", entry.Name, time.Now().UnixNano()/int64(time.Millisecond)%1000000)
	pod, err := buildStoragePodSpec(podName, namespace, entry, input)
	if err != nil {
		return err
	}

	common.LogVerboseQuiet(fmt.Sprintf("Creating exec pod %s/%s", namespace, podName))
	if _, err := clientset.Client.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("error creating exec pod %s/%s: %w", namespace, podName, err)
	}

	defer func() {
		bg := context.Background()
		gracePeriod := int64(0)
		_ = clientset.Client.CoreV1().Pods(namespace).Delete(bg, podName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	}()

	if err := waitForPodRunning(ctx, clientset, namespace, podName, 60*time.Second); err != nil {
		return err
	}

	command := input.Command
	if len(command) == 0 {
		command = []string{"sh", "-c", "command -v bash >/dev/null 2>&1 && exec bash || exec sh"}
	}

	execErr := clientset.ExecCommand(ctx, ExecCommandInput{
		Name:          podName,
		Namespace:     namespace,
		ContainerName: "exec",
		Command:       command,
	})
	if execErr == nil {
		return nil
	}

	var coded utilexec.CodeExitError
	if errors.As(execErr, &coded) {
		os.Exit(coded.Code)
	}
	return execErr
}

// buildStoragePodSpec assembles the throwaway Pod that mounts the PVC
// for storage:exec. Split out so the unit test can verify the spec
// without hitting a cluster.
func buildStoragePodSpec(name string, namespace string, entry *storage.Entry, input StorageExecInput) (*corev1.Pod, error) {
	uid, err := resolveStorageExecUID(entry.Chown, input.AsUser)
	if err != nil {
		return nil, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "dokku",
				"dokku.com/storage-entry":      entry.Name,
				"dokku.com/purpose":            "storage-exec",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: entry.Name,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "exec",
					Image:   input.Image,
					Command: []string{"sleep", "infinity"},
					Stdin:   input.Interactive,
					TTY:     input.Tty,
					VolumeMounts: []corev1.VolumeMount{
						{Name: "data", MountPath: "/data"},
					},
				},
			},
		},
	}

	if uid != nil {
		pod.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser:  uid,
			RunAsGroup: uid,
			FSGroup:    uid,
		}
	}

	return pod, nil
}

// resolveStorageExecUID picks the UID to run the exec Pod as. Returns
// nil to mean "leave the image's default user in place".
func resolveStorageExecUID(chown string, asUser string) (*int64, error) {
	if asUser != "" {
		n, err := strconv.ParseInt(asUser, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("--as-user must be a numeric uid, got %q", asUser)
		}
		return &n, nil
	}
	if chown == "" || chown == "false" {
		return nil, nil
	}
	uidStr, err := storage.ResolveChownID(chown)
	if err != nil {
		return nil, err
	}
	if uidStr == "" || uidStr == "false" {
		return nil, nil
	}
	n, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse resolved chown UID %q: %w", uidStr, err)
	}
	return &n, nil
}

// waitForPodRunning polls the Pod until its phase is Running, surfacing
// the structured Kubernetes API state on failure (image pull failures,
// scheduling problems, etc.) rather than a generic timeout.
func waitForPodRunning(ctx context.Context, clientset KubernetesClient, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		pod, err := clientset.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error reading exec pod %s/%s: %w", namespace, name, err)
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			return nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return fmt.Errorf("exec pod %s/%s reached terminal phase %s before exec could attach", namespace, name, pod.Status.Phase)
		}

		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting == nil {
				continue
			}
			reason := cs.State.Waiting.Reason
			switch reason {
			case "ImagePullBackOff", "ErrImagePull", "CreateContainerConfigError", "InvalidImageName":
				return fmt.Errorf("exec pod %s/%s container %q failed to start: %s: %s", namespace, name, cs.Name, reason, cs.State.Waiting.Message)
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("exec pod %s/%s did not reach Running within %s (current phase: %s)", namespace, name, timeout, pod.Status.Phase)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
