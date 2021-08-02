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

func incrementTagVersion(appName string) (string, error) {
	tag := common.PropertyGet("registry", appName, "tag-version")
	if tag == "" {
		tag = "0"
	}

	version, err := strconv.Atoi(tag)
	if err != nil {
		return "", fmt.Errorf("Unable to convert existing tag version (%s) to integer: %v", tag, err)
	}

	version++
	common.LogVerboseQuiet(fmt.Sprintf("Bumping tag to %d", version))
	if err = common.PropertyWrite("registry", appName, "tag-version", strconv.Itoa(version)); err != nil {
		return "", err
	}

	return strconv.Itoa(version), nil
}

func pushToRegistry(appName string, tag string) error {
	common.LogVerboseQuiet("Retrieving image info for app")

	registryServer := getRegistryServerForApp(appName)
	imageRepo := reportComputedImageRepo(appName)
	imageTag, _ := common.GetRunningImageTag(appName)
	image := common.GetAppImageName(appName, imageTag, "")
	imageID, _ := common.DockerInspect(image, "{{ .Id }}")

	common.LogVerboseQuiet(fmt.Sprintf("Tagging $IMAGE_REPO:%s in registry format", tag))
	if !dockerTag(imageID, fmt.Sprintf("%s%s:%s", registryServer, imageRepo, tag)) {
		// TODO: better error
		return errors.New("Unable to tag image")
	}

	if !dockerTag(imageID, fmt.Sprintf("%s:%s", imageRepo, tag)) {
		// TODO: better error
		return errors.New("Unable to tag image")
	}

	// fn-registry-create-repository "$APP" "$DOKKU_REGISTRY_SERVER" "$IMAGE_REPO"

	common.LogVerboseQuiet("Pushing $IMAGE_REPO:$TAG")
	if !dockerPush(fmt.Sprintf("%s%s:%s", registryServer, imageRepo, tag)) {
		// TODO: better error
		return errors.New("Unable to push image")
	}

	common.LogVerboseQuiet("Cleaning up")
	// fn-registry-image-cleanup "$APP" "$DOKKU_REGISTRY_SERVER" "$IMAGE_REPO" "$IMAGE_TAG" "$TAG"

	common.LogVerboseQuiet("Image $IMAGE_REPO:$TAG pushed")
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
