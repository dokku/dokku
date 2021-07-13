package registry

import (
	"fmt"
	"strconv"
	"strings"

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

func pushToRegistry(appName string, imageTag string) error {
	common.LogVerboseQuiet("Retrieving image info for app")

	// DOKKU_REGISTRY_SERVER=$(fn-registry-remote-repository "$APP")
	// IMAGE_REPO=$(fn-registry-image-repo "$APP")
	// IMAGE_TAG="$(get_running_image_tag "$APP")"
	// IMAGE=$(get_app_image_name "$APP" "$IMAGE_TAG")
	// IMAGE_ID=$(docker inspect --format '{{ .Id }}' "$IMAGE")

	// dokku_log_verbose_quiet "Tagging $IMAGE_REPO:$TAG in registry format"
	// docker tag "$IMAGE_ID" "${DOKKU_REGISTRY_SERVER}${IMAGE_REPO}:${TAG}"
	// docker tag "$IMAGE_ID" "${IMAGE_REPO}:${TAG}"

	// fn-registry-create-repository "$APP" "$DOKKU_REGISTRY_SERVER" "$IMAGE_REPO"

	common.LogVerboseQuiet("Pushing $IMAGE_REPO:$TAG")
	// docker push "${DOKKU_REGISTRY_SERVER}${IMAGE_REPO}:${TAG}"

	common.LogVerboseQuiet("Cleaning up")
	// fn-registry-image-cleanup "$APP" "$DOKKU_REGISTRY_SERVER" "$IMAGE_REPO" "$IMAGE_TAG" "$TAG"

	common.LogVerboseQuiet("Image $IMAGE_REPO:$TAG pushed")
	return nil
}
