package scheduler_k3s

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// ImageMetadata holds the image-derived values a chart or job needs that would
// otherwise require the image to be present on the local docker daemon.
type ImageMetadata struct {
	SourceType string
	WorkingDir string
}

// resolveImageMetadata determines the builder type and working directory for an
// image, preferring a local inspect and falling back to values carried by the
// app's current Helm release.
//
// k3s workloads run in the cluster and kubelet pulls the image itself, so the
// dokku host is free to reap its local copy while the app keeps running - the
// registry plugin's own imageCleanup does exactly that. Without a fallback,
// every later deploy of such an app would fail. A fallback alone is not enough
// either: common.IsImageCnbBased, common.IsImageHerokuishBased and
// common.GetWorkingDir all report false/"" when the inspect errors, so a
// missing image would otherwise be silently misread as a dockerfile app with no
// working directory, dropping the herokuish /start wrapper from the start
// command.
func resolveImageMetadata(appName string, image string, fallback *ImageMetadata) (ImageMetadata, error) {
	if common.VerifyImage(image) {
		metadata := ImageMetadata{SourceType: "dockerfile"}
		if common.IsImageCnbBased(image) {
			metadata.SourceType = "pack"
		} else if common.IsImageHerokuishBased(image, appName) {
			metadata.SourceType = "herokuish"
		}

		metadata.WorkingDir = common.GetWorkingDir(appName, image)
		return metadata, nil
	}

	if fallback != nil && fallback.SourceType != "" {
		return *fallback, nil
	}

	return ImageMetadata{}, fmt.Errorf("App image (%s) not found locally and no deployed release to read image metadata from", image)
}

// releaseValues reads an app's current Helm release values, which every deploy
// writes and which therefore make the cluster the authoritative record of how
// the running workload was built. Returns nil when there is no readable
// release, leaving callers to treat the values as unavailable rather than
// failing - a first deploy has no release yet. Existence is checked up front so
// a never-deployed app does not surface a not-found through the agent's logger.
func releaseValues(helmAgent *HelmAgent, releaseName string) map[string]interface{} {
	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil || !exists {
		return nil
	}

	values, err := helmAgent.GetValues(releaseName)
	if err != nil {
		return nil
	}

	return values
}

// imageMetadataFromValues extracts image metadata from a set of Helm release values.
func imageMetadataFromValues(values map[string]interface{}) *ImageMetadata {
	globalValues, ok := values["global"].(map[string]interface{})
	if !ok {
		return nil
	}

	imageValues, ok := globalValues["image"].(map[string]interface{})
	if !ok {
		return nil
	}

	sourceType, ok := imageValues["type"].(string)
	if !ok || sourceType == "" {
		return nil
	}

	workingDir, _ := imageValues["working_dir"].(string)
	return &ImageMetadata{
		SourceType: sourceType,
		WorkingDir: workingDir,
	}
}

// processDeploymentIDsFromValues extracts the per-process deployment ids
// recorded in an app's current Helm release, keyed by process type.
//
// A targeted restart must leave the untargeted processes' pod templates byte
// identical, otherwise Kubernetes rolls them too. Their ids therefore have to
// come from what is already deployed rather than being regenerated. Releases
// written before per-process ids existed only carry a global id, so that is
// used as the per-process default.
func processDeploymentIDsFromValues(values map[string]interface{}) map[string]string {
	deploymentIDs := map[string]string{}

	globalDeploymentID := ""
	if globalValues, ok := values["global"].(map[string]interface{}); ok {
		globalDeploymentID, _ = globalValues["deployment_id"].(string)
	}

	processes, ok := values["processes"].(map[string]interface{})
	if !ok {
		return deploymentIDs
	}

	for processType, rawValues := range processes {
		processValues, ok := rawValues.(map[string]interface{})
		if !ok {
			continue
		}

		deploymentID, _ := processValues["deployment_id"].(string)
		if deploymentID == "" {
			deploymentID = globalDeploymentID
		}
		if deploymentID != "" {
			deploymentIDs[processType] = deploymentID
		}
	}

	return deploymentIDs
}

// resolveProcessDeploymentID picks the deployment id for a single process type.
// Every process type gets the fresh id unless a different one was targeted for
// restart and a prior id is known for this one, in which case holding that id
// steady leaves the process's pod template untouched so Kubernetes does not roll
// it. A targeted restart against an app with no prior release rolls everything,
// which is correct - there is nothing running to preserve.
func resolveProcessDeploymentID(processType string, deploymentID int64, restartProcessType string, priorDeploymentIDs map[string]string) string {
	if restartProcessType != "" && restartProcessType != processType {
		if priorDeploymentID := priorDeploymentIDs[processType]; priorDeploymentID != "" {
			return priorDeploymentID
		}
	}

	return fmt.Sprint(deploymentID)
}
