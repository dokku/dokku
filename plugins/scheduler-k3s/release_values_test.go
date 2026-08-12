package scheduler_k3s

import (
	"strings"
	"testing"
)

func TestImageMetadataFromValues(t *testing.T) {
	t.Run("reads the builder type and working directory", func(t *testing.T) {
		metadata := imageMetadataFromValues(map[string]interface{}{
			"global": map[string]interface{}{
				"image": map[string]interface{}{
					"name":        "registry.example.com/dokku/myapp:5",
					"type":        "herokuish",
					"working_dir": "/app",
				},
			},
		})

		if metadata == nil {
			t.Fatal("imageMetadataFromValues() = nil, want metadata")
		}
		if metadata.SourceType != "herokuish" {
			t.Errorf("SourceType = %q, want herokuish", metadata.SourceType)
		}
		if metadata.WorkingDir != "/app" {
			t.Errorf("WorkingDir = %q, want /app", metadata.WorkingDir)
		}
	})

	t.Run("allows an empty working directory", func(t *testing.T) {
		metadata := imageMetadataFromValues(map[string]interface{}{
			"global": map[string]interface{}{
				"image": map[string]interface{}{"type": "dockerfile"},
			},
		})

		if metadata == nil {
			t.Fatal("imageMetadataFromValues() = nil, want metadata")
		}
		if metadata.WorkingDir != "" {
			t.Errorf("WorkingDir = %q, want empty", metadata.WorkingDir)
		}
	})

	t.Run("nil when the release carries no usable type", func(t *testing.T) {
		cases := map[string]map[string]interface{}{
			"no global key": {},
			"no image key":  {"global": map[string]interface{}{}},
			"no type key": {
				"global": map[string]interface{}{"image": map[string]interface{}{"name": "myapp:latest"}},
			},
			"empty type": {
				"global": map[string]interface{}{"image": map[string]interface{}{"type": ""}},
			},
			"non-string type": {
				"global": map[string]interface{}{"image": map[string]interface{}{"type": 3}},
			},
		}

		for name, values := range cases {
			t.Run(name, func(t *testing.T) {
				if metadata := imageMetadataFromValues(values); metadata != nil {
					t.Errorf("imageMetadataFromValues() = %+v, want nil", metadata)
				}
			})
		}
	})
}

func TestResolveImageMetadata(t *testing.T) {
	// These cases all rely on the image being absent from the local docker
	// daemon, which is what a reaped k3s image looks like.
	const missingImage = "registry.example.invalid/dokku/does-not-exist:0"

	t.Run("uses the fallback when the image is not present locally", func(t *testing.T) {
		metadata, err := resolveImageMetadata("myapp", missingImage, &ImageMetadata{
			SourceType: "herokuish",
			WorkingDir: "/app",
		})
		if err != nil {
			t.Fatalf("resolveImageMetadata() error = %v, want nil", err)
		}
		if metadata.SourceType != "herokuish" {
			t.Errorf("SourceType = %q, want herokuish", metadata.SourceType)
		}
		if metadata.WorkingDir != "/app" {
			t.Errorf("WorkingDir = %q, want /app", metadata.WorkingDir)
		}
	})

	t.Run("errors when there is no image and no fallback", func(t *testing.T) {
		_, err := resolveImageMetadata("myapp", missingImage, nil)
		if err == nil {
			t.Fatal("resolveImageMetadata() error = nil, want an error")
		}
		if !strings.Contains(err.Error(), missingImage) {
			t.Errorf("error %q does not name the image", err.Error())
		}
	})

	t.Run("ignores a fallback carrying no builder type", func(t *testing.T) {
		if _, err := resolveImageMetadata("myapp", missingImage, &ImageMetadata{WorkingDir: "/app"}); err == nil {
			t.Fatal("resolveImageMetadata() error = nil, want an error")
		}
	})
}

func TestProcessDeploymentIDsFromValues(t *testing.T) {
	t.Run("reads per-process ids", func(t *testing.T) {
		deploymentIDs := processDeploymentIDsFromValues(map[string]interface{}{
			"global": map[string]interface{}{"deployment_id": "100"},
			"processes": map[string]interface{}{
				"web":    map[string]interface{}{"deployment_id": "200"},
				"worker": map[string]interface{}{"deployment_id": "300"},
			},
		})

		if deploymentIDs["web"] != "200" {
			t.Errorf("web = %q, want 200", deploymentIDs["web"])
		}
		if deploymentIDs["worker"] != "300" {
			t.Errorf("worker = %q, want 300", deploymentIDs["worker"])
		}
	})

	t.Run("falls back to the global id for releases predating per-process ids", func(t *testing.T) {
		deploymentIDs := processDeploymentIDsFromValues(map[string]interface{}{
			"global": map[string]interface{}{"deployment_id": "100"},
			"processes": map[string]interface{}{
				"web":    map[string]interface{}{"replicas": 1},
				"worker": map[string]interface{}{"deployment_id": "300"},
			},
		})

		if deploymentIDs["web"] != "100" {
			t.Errorf("web = %q, want the global 100", deploymentIDs["web"])
		}
		if deploymentIDs["worker"] != "300" {
			t.Errorf("worker = %q, want 300", deploymentIDs["worker"])
		}
	})

	t.Run("empty when there is nothing to read", func(t *testing.T) {
		cases := map[string]map[string]interface{}{
			"no values":       {},
			"no processes":    {"global": map[string]interface{}{"deployment_id": "100"}},
			"no ids anywhere": {"processes": map[string]interface{}{"web": map[string]interface{}{"replicas": 1}}},
		}

		for name, values := range cases {
			t.Run(name, func(t *testing.T) {
				if deploymentIDs := processDeploymentIDsFromValues(values); len(deploymentIDs) != 0 {
					t.Errorf("processDeploymentIDsFromValues() = %v, want empty", deploymentIDs)
				}
			})
		}
	})
}

func TestResolveProcessDeploymentID(t *testing.T) {
	priorDeploymentIDs := map[string]string{"web": "100", "worker": "100"}

	t.Run("an untargeted restart rolls every process", func(t *testing.T) {
		for _, processType := range []string{"web", "worker"} {
			if got := resolveProcessDeploymentID(processType, 200, "", priorDeploymentIDs); got != "200" {
				t.Errorf("resolveProcessDeploymentID(%q) = %q, want 200", processType, got)
			}
		}
	})

	t.Run("a targeted restart holds the other processes steady", func(t *testing.T) {
		if got := resolveProcessDeploymentID("web", 200, "web", priorDeploymentIDs); got != "200" {
			t.Errorf("targeted process = %q, want 200", got)
		}
		if got := resolveProcessDeploymentID("worker", 200, "web", priorDeploymentIDs); got != "100" {
			t.Errorf("untargeted process = %q, want the prior 100", got)
		}
	})

	t.Run("a process with no prior id takes the fresh one", func(t *testing.T) {
		if got := resolveProcessDeploymentID("worker", 200, "web", map[string]string{}); got != "200" {
			t.Errorf("resolveProcessDeploymentID() = %q, want 200", got)
		}
	})
}
