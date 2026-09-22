package scheduler_k3s

import (
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/storage"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func storageEntryK3s(name string) *storage.Entry {
	return &storage.Entry{Name: name, Scheduler: storage.SchedulerK3s}
}

func storageAttachment(containerPath string, subpath string) *storage.Attachment {
	return &storage.Attachment{ContainerPath: containerPath, Subpath: subpath}
}

// renderProcessChart renders one of the chart templates against the given
// process values, going through YAML the same way a real deploy does. That
// round trip is the point: it exercises the struct's yaml tags against the
// keys the template actually reads, so a rename on one side without the other
// fails here instead of at helm upgrade time.
func renderProcessChart(t *testing.T, templateName string, processName string, values ProcessValues) string {
	t.Helper()

	appValues := AppValues{
		Global: GlobalValues{
			AppName:      "demo",
			DeploymentID: "1",
			Namespace:    "default",
			Image: GlobalImage{
				Name: "dokku/demo:latest",
				Type: "herokuish",
			},
		},
		Processes: map[string]ProcessValues{processName: values},
	}

	encoded, err := yaml.Marshal(appValues)
	if err != nil {
		t.Fatalf("marshal values: %v", err)
	}
	rendered := map[string]interface{}{}
	if err := yaml.Unmarshal(encoded, &rendered); err != nil {
		t.Fatalf("unmarshal values: %v", err)
	}

	templateBody, err := templates.ReadFile("templates/chart/" + templateName)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	helpers, err := templates.ReadFile("templates/chart/_helpers.tpl")
	if err != nil {
		t.Fatalf("read helpers: %v", err)
	}

	c := &chart.Chart{
		Metadata: &chart.Metadata{APIVersion: chart.APIVersionV2, Name: "demo", Version: "0.0.1"},
		Templates: []*chart.File{
			{Name: "templates/" + templateName, Data: templateBody},
			{Name: "templates/_helpers.tpl", Data: helpers},
		},
	}

	top, err := chartutil.ToRenderValues(c, rendered, chartutil.ReleaseOptions{Name: "demo", Namespace: "default"}, nil)
	if err != nil {
		t.Fatalf("render values: %v", err)
	}
	out, err := engine.Render(c, top)
	if err != nil {
		t.Fatalf("render chart: %v", err)
	}

	return out["demo/templates/"+templateName]
}

// podSpecFrom decodes the rendered manifest and returns the pod spec at the
// given path, so the assertions below read the same structure the API server
// would rather than counting indentation.
func podSpecFrom(t *testing.T, manifest string, path ...string) map[string]interface{} {
	t.Helper()

	decoder := yaml.NewDecoder(strings.NewReader(manifest))
	doc := map[string]interface{}{}
	for {
		candidate := map[string]interface{}{}
		if err := decoder.Decode(&candidate); err != nil {
			break
		}
		if len(candidate) > 0 {
			doc = candidate
			break
		}
	}
	if len(doc) == 0 {
		t.Fatalf("no document decoded from manifest:\n%s", manifest)
	}

	current := doc
	for _, key := range path {
		next, ok := current[key].(map[string]interface{})
		if !ok {
			t.Fatalf("key %q missing while walking %v in manifest:\n%s", key, path, manifest)
		}
		current = next
	}
	return current
}

// volumeNamesFrom returns the pod-level volume names, and mountsFrom the
// container-level mounts, of a decoded pod spec.
func volumeNamesFrom(t *testing.T, podSpec map[string]interface{}) []string {
	t.Helper()

	names := []string{}
	volumes, _ := podSpec["volumes"].([]interface{})
	for _, entry := range volumes {
		volume, ok := entry.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected volume entry %#v", entry)
		}
		names = append(names, volume["name"].(string))
	}
	return names
}

func mountsFrom(t *testing.T, podSpec map[string]interface{}) []map[string]interface{} {
	t.Helper()

	containers, _ := podSpec["containers"].([]interface{})
	if len(containers) == 0 {
		t.Fatalf("no containers in pod spec %#v", podSpec)
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected container entry %#v", containers[0])
	}

	mounts := []map[string]interface{}{}
	raw, _ := container["volumeMounts"].([]interface{})
	for _, entry := range raw {
		mount, ok := entry.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected volumeMount entry %#v", entry)
		}
		mounts = append(mounts, mount)
	}
	return mounts
}

// TestDeploymentRendersOneVolumePerNameAndOneMountPerPath is the regression
// for a storage entry mounted at two container paths. Kubernetes requires pod
// volume names to be unique, so the pod must declare the PVC once and bind it
// twice - emitting it twice is rejected with "Duplicate value".
func TestDeploymentRendersOneVolumePerNameAndOneMountPerPath(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry:      storageEntryK3s("demo-pvc"),
			Attachment: storageAttachment("/one", ""),
		},
		{
			Entry:      storageEntryK3s("demo-pvc"),
			Attachment: storageAttachment("/two", "sub"),
		},
	}

	manifest := renderProcessChart(t, "deployment.yaml", "web", ProcessValues{
		ProcessType: ProcessType_Worker,
		Replicas:    1,
		Volumes:     ToProcessVolumes(pairs, "web"),
	})

	podSpec := podSpecFrom(t, manifest, "spec", "template", "spec")

	names := volumeNamesFrom(t, podSpec)
	if len(names) != 1 || names[0] != "demo-pvc" {
		t.Fatalf("expected a single pod volume named demo-pvc, got %v\n%s", names, manifest)
	}

	mounts := mountsFrom(t, podSpec)
	if len(mounts) != 2 {
		t.Fatalf("expected 2 volume mounts, got %d\n%s", len(mounts), manifest)
	}
	if mounts[0]["mountPath"] != "/one" || mounts[1]["mountPath"] != "/two" {
		t.Errorf("expected mounts at /one and /two, got %v and %v", mounts[0]["mountPath"], mounts[1]["mountPath"])
	}
	if mounts[1]["subPath"] != "sub" {
		t.Errorf("expected subPath sub on the second mount, got %v", mounts[1]["subPath"])
	}
	if !strings.Contains(manifest, "claimName: demo-pvc") {
		t.Errorf("expected the PVC to be referenced, got:\n%s", manifest)
	}
}

// TestCronJobRendersOneVolumePerName covers the same rule in the cron-job
// template, which renders its volumes from the same two lists.
func TestCronJobRendersOneVolumePerName(t *testing.T) {
	pairs := []AppMountPair{
		{
			Entry:      storageEntryK3s("demo-pvc"),
			Attachment: storageAttachment("/one", ""),
		},
		{
			Entry:      storageEntryK3s("demo-pvc"),
			Attachment: storageAttachment("/two", ""),
		},
	}

	manifest := renderProcessChart(t, "cron-job.yaml", "cron-1", ProcessValues{
		ProcessType: ProcessType_Cron,
		Replicas:    1,
		Cron: ProcessCron{
			ID:       "abc123",
			Hash:     "abc123",
			Schedule: "*/5 * * * *",
			Suffix:   "aaaaa",
		},
		Volumes: ToProcessVolumes(pairs, ""),
	})

	podSpec := podSpecFrom(t, manifest, "spec", "jobTemplate", "spec", "template", "spec")

	if names := volumeNamesFrom(t, podSpec); len(names) != 1 || names[0] != "demo-pvc" {
		t.Fatalf("expected a single pod volume named demo-pvc, got %v\n%s", names, manifest)
	}
	if mounts := mountsFrom(t, podSpec); len(mounts) != 2 {
		t.Fatalf("expected 2 volume mounts, got %d\n%s", len(mounts), manifest)
	}

	// A cron expression starting with `*` is not a valid plain YAML scalar, so
	// the schedule has to render quoted or the whole manifest fails to parse.
	spec := podSpecFrom(t, manifest, "spec")
	if spec["schedule"] != "*/5 * * * *" {
		t.Errorf("expected the schedule to survive rendering, got %v", spec["schedule"])
	}
}
