package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/dokku/dokku/plugins/common"
)

const registryConfigDir = "/var/lib/dokku/config/registry"

// dockerIndexServer is the auths key docker stores Docker Hub credentials under
const dockerIndexServer = "https://index.docker.io/v1/"

// GetAppRegistryConfigDir returns the per-app registry config directory
func GetAppRegistryConfigDir(appName string) string {
	return filepath.Join(registryConfigDir, appName)
}

// GetAppRegistryConfigPath returns the path to per-app docker config.json
func GetAppRegistryConfigPath(appName string) string {
	return filepath.Join(GetAppRegistryConfigDir(appName), "config.json")
}

func GetComputedAppRegistryConfigDir(appName string) string {
	if HasAppRegistryAuth(appName) {
		return GetAppRegistryConfigDir(appName)
	}

	return GetGlobalRegistryConfigDir()
}

// GetGlobalRegistryConfigDir returns the docker config directory global logins write to
func GetGlobalRegistryConfigDir() string {
	return filepath.Join(os.Getenv("DOKKU_ROOT"), ".docker")
}

// GetGlobalRegistryConfigPath returns the path to the global docker config.json
func GetGlobalRegistryConfigPath() string {
	return filepath.Join(GetGlobalRegistryConfigDir(), "config.json")
}

// HasAppRegistryAuth checks if an app has registry credentials configured
func HasAppRegistryAuth(appName string) bool {
	configPath := GetAppRegistryConfigPath(appName)
	if !common.FileExists(configPath) {
		return false
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	var config map[string]interface{}
	if err := json.Unmarshal(content, &config); err != nil {
		return false
	}

	auths, ok := config["auths"].(map[string]interface{})
	return ok && len(auths) > 0
}

// GetDockerConfigArgs returns docker --config arguments if per-app config exists
func GetDockerConfigArgs(appName string) []string {
	if appName == "" || !HasAppRegistryAuth(appName) {
		return []string{}
	}
	return []string{"--config", GetAppRegistryConfigDir(appName)}
}

func getImageRepoFromTemplate(appName string) (string, error) {
	imageRepoTemplate := strings.TrimSpace(reportImageRepoTemplate(appName))
	if imageRepoTemplate == "" {
		imageRepoTemplate = reportGlobalImageRepoTemplate(appName)
	}
	if imageRepoTemplate == "" {
		return "", nil
	}

	tmpl, err := template.New("template").Parse(imageRepoTemplate)
	if err != nil {
		return "", fmt.Errorf("Unable to parse image-repo-template: %w", err)
	}

	type templateData struct {
		AppName string
	}
	data := templateData{AppName: appName}

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, data); err != nil {
		return "", fmt.Errorf("Unable to execute image-repo-template: %w", err)
	}

	return strings.TrimSpace(doc.String()), nil
}

func getRegistryServerForApp(appName string) string {
	value := common.PropertyGet("registry", appName, "server")
	if value == "" {
		value = common.PropertyGet("registry", "--global", "server")
	}
	value = strings.TrimSpace(value)

	value = strings.TrimSuffix(value, "/")
	if value == "hub.docker.com" || value == "docker.io" {
		value = ""
	}

	if value != "" {
		value = value + "/"
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

	tag = strings.TrimSpace(tag)
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

func getRegistryPushExtraTagsForApp(appName string) string {
	return reportComputedPushExtraTags(appName)
}

func pushToRegistry(appName string, tag int, imageID string, imageRepo string) error {
	common.LogVerboseQuiet("Retrieving image info for app")

	registryServer := getRegistryServerForApp(appName)
	imageTag, _ := common.GetRunningImageTag(appName, "")

	fullImage := fmt.Sprintf("%s%s:%d", registryServer, imageRepo, tag)

	common.LogVerboseQuiet(fmt.Sprintf("Tagging %s:%d in registry format", imageRepo, tag))
	if err := dockerTag(imageID, fullImage); err != nil {
		return fmt.Errorf("unable to tag image %s as %s: %w", imageID, fullImage, err)
	}

	if err := dockerTag(imageID, fmt.Sprintf("%s:%d", imageRepo, tag)); err != nil {
		return fmt.Errorf("unable to tag image %s as %s:%d: %w", imageID, imageRepo, tag, err)
	}

	extraTags := getRegistryPushExtraTagsForApp(appName)
	if extraTags != "" {
		extraTagsArray := strings.Split(extraTags, ",")
		for _, extraTag := range extraTagsArray {
			extraTagImage := fmt.Sprintf("%s%s:%s", registryServer, imageRepo, extraTag)
			common.LogVerboseQuiet(fmt.Sprintf("Tagging %s as %s in registry format", imageRepo, extraTag))
			if err := dockerTag(imageID, extraTagImage); err != nil {
				return fmt.Errorf("unable to tag image %s as %s: %w", imageID, extraTagImage, err)
			}
			defer func() {
				common.LogVerboseQuiet(fmt.Sprintf("Untagging extra tag %s", extraTag))
				if err := common.RemoveImages([]string{extraTagImage}); err != nil {
					common.LogWarn(fmt.Sprintf("Unable to untag extra tag %s: %s", extraTag, err.Error()))
				}
			}()
			common.LogVerboseQuiet(fmt.Sprintf("Pushing %s", extraTagImage))
			if err := dockerPush(appName, extraTagImage); err != nil {
				return fmt.Errorf("unable to push image with %s tag: %w", extraTag, err)
			}
		}
	}

	common.LogVerboseQuiet(fmt.Sprintf("Pushing %s", fullImage))
	if err := dockerPush(appName, fullImage); err != nil {
		return fmt.Errorf("unable to push image %s: %w", fullImage, err)
	}

	// Only clean up when the scheduler is not docker-local
	// other schedulers do not retire local images
	if common.GetAppScheduler(appName) != "docker-local" {
		common.LogVerboseQuiet("Cleaning up")
		imageCleanup(appName, fmt.Sprintf("%s%s", registryServer, imageRepo), imageTag, tag)
		if fmt.Sprintf("%s%s", registryServer, imageRepo) != imageRepo {
			imageCleanup(appName, imageRepo, imageTag, tag)
		}
	}

	common.LogVerboseQuiet(fmt.Sprintf("Image %s pushed", fullImage))
	return nil
}

func dockerTag(imageID string, imageTag string) error {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     common.DockerBin(),
		Args:        []string{"image", "tag", imageID, imageTag},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("docker tag command failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("docker tag command exited with code %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

func dockerPush(appName string, imageTag string) error {
	args := GetDockerConfigArgs(appName)
	args = append(args, "image", "push", imageTag)
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     common.DockerBin(),
		Args:        args,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("docker image push command failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("docker image push command exited with code %d: %s", result.ExitCode, result.Stderr)
	}
	return nil
}

func imageCleanup(appName string, imageRepo string, imageTag string, tag int) {
	// # keep last two images in place
	oldTag := tag - 2
	tenImagesAgoTag := tag - 12

	imagesToRemove := []string{}
	for oldTag > 0 {
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

// normalizeRegistryServer applies the dokku-level aliases for a registry server
func normalizeRegistryServer(server string) string {
	if server == "hub.docker.com" || server == "docker.com" {
		return "docker.io"
	}

	return server
}

// convertToHostname strips any scheme and path from a registry server, matching
// the rule docker applies before storing or looking up a credential
func convertToHostname(server string) string {
	stripped := server
	if strings.HasPrefix(stripped, "http://") {
		stripped = strings.TrimPrefix(stripped, "http://")
	} else if strings.HasPrefix(stripped, "https://") {
		stripped = strings.TrimPrefix(stripped, "https://")
	}

	return strings.SplitN(stripped, "/", 2)[0]
}

// dockerAuthKey returns the key docker uses for a registry server in the auths
// map of a config.json. Docker Hub is stored under the index server address
// rather than the hostname the user logged in with, and is also the only server
// docker treats that way: registry-1.docker.io gets a key of its own.
func dockerAuthKey(server string) string {
	hostname := convertToHostname(normalizeRegistryServer(server))
	if hostname == "" || hostname == "docker.io" || hostname == "index.docker.io" {
		return dockerIndexServer
	}

	return hostname
}
