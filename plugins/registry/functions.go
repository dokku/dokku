package registry

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/dokku/dokku/plugins/common"
)

func getRegistryServerForApp(appName string) string {
	value := common.PropertyGet("registry", appName, "server")
	if value == "" {
		value = common.PropertyGet("registry", "--global", "server")
	}

	value = strings.TrimSuffix(value, "/") + "/"
	if value == "hub.docker.com/" || value == "docker.io/" {
		value = ""
	}

	return value
}

func isPushEnabled(appName string) bool {
	return reportComputedPushOnRelease(appName) == "true"
}

func incrementTagVersion(appName string) (int, error) {
	tag := common.PropertyGet("registry", appName, "tag-version")
	if tag == "" {
		tag = "0"
	}

	version, err := strconv.Atoi(tag)
	if err != nil {
		return 0, fmt.Errorf("Unable to convert existing tag version (%s) to integer: %v", tag, err)
	}

	version++
	common.LogVerboseQuiet(fmt.Sprintf("Bumping tag to %d", version))
	if err = common.PropertyWrite("registry", appName, "tag-version", strconv.Itoa(version)); err != nil {
		return 0, err
	}

	return version, nil
}

func pushToRegistry(appName string, tag int, imageID string, imageRepo string) error {
	common.LogVerboseQuiet("Retrieving image info for app")

	registryServer := getRegistryServerForApp(appName)
	imageTag, _ := common.GetRunningImageTag(appName)

	fullImage := fmt.Sprintf("%s%s:%d", registryServer, imageRepo, tag)

	common.LogVerboseQuiet(fmt.Sprintf("Tagging %s:%d in registry format", imageRepo, tag))
	if !dockerTag(imageID, fullImage) {
		// TODO: better error
		return errors.New("Unable to tag image")
	}

	if !dockerTag(imageID, fmt.Sprintf("%s:%d", imageRepo, tag)) {
		// TODO: better error
		return errors.New("Unable to tag image")
	}

	// For the future, we should also add the ability to create the remote repository
	// This is only really important for registries that do not support creation on push
	// Examples include AWS and Quay.io

	common.LogVerboseQuiet(fmt.Sprintf("Pushing %s", fullImage))
	if !dockerPush(fullImage) {
		// TODO: better error
		return errors.New("Unable to push image")
	}

	common.LogVerboseQuiet("Cleaning up")
	imageCleanup(appName, fmt.Sprintf("%s%s", registryServer, imageRepo), imageTag, tag)
	imageCleanup(appName, imageRepo, imageTag, tag)

	common.LogVerboseQuiet(fmt.Sprintf("Image %s pushed", fullImage))
	return nil
}

func dockerTag(imageID string, imageTag string) bool {
	cmd := sh.Command(common.DockerBin(), "image", "tag", imageID, imageTag)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func dockerPush(imageTag string) bool {
	cmd := sh.Command(common.DockerBin(), "image", "push", imageTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func imageCleanup(appName string, imageRepo string, imageTag string, tag int) {
	// # keep last two images in place
	oldTag := tag - 1
	tenImagesAgoTag := tag - 12

	imagesToRemove := []string{}
	for oldTag > 0 {
		common.LogInfo1(fmt.Sprintf("Removing image: %s:%d", imageRepo, oldTag))
		imagesToRemove = append(imagesToRemove, fmt.Sprintf("%s:%d", imageRepo, oldTag))
		oldTag = oldTag - 1
		if tenImagesAgoTag == oldTag {
			break
		}
	}

	imageIDs, _ := common.ListDanglingImages(appName)
	imagesToRemove = append(imagesToRemove, imageIDs...)
	common.RemoveImages(imagesToRemove)
}
