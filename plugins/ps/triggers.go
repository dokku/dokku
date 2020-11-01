package ps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
	sh "github.com/codeskyblue/go-sh"
)

func TriggerAppRestart(appName string) error {
	return Restart(appName)
}

func TriggerCorePostDeploy(appName string) error {
	if err := removeProcfile(appName); err != nil {
		return err
	}

	entries := map[string]string{
		"DOKKU_APP_RESTORE": "1",
	}

	return common.SuppressOutput(func() error {
		return config.SetMany(appName, entries, false)
	})
}

func TriggerInstall() error {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	apps, err := common.DokkuApps()
	if err != nil {
		return nil
	}

	for _, appName := range apps {
		policies, err := getRestartPolicy(appName)
		if err != nil {
			return err
		}

		if len(policies) != 0 {
			continue
		}

		if err := dockeroptions.AddDockerOptionToPhases(appName, []string{"deploy"}, "--restart=on-failure:10"); err != nil {
			common.LogWarn(err.Error())
		}
	}

	return nil
}

func TriggerPostAppClone(oldAppName string, newAppName string) error {
	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

func TriggerPostAppRename(oldAppName string, newAppName string) error {
	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

func TriggerPostCreate(appName string) error {
	if err := dockeroptions.AddDockerOptionToPhases(appName, []string{"deploy"}, "--restart=on-failure:10"); err != nil {
		return err
	}

	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps", appName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the ps properties for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("ps", appName)
}

func TriggerPostExtract(appName string, tempWorkDir string) error {
	procfile := filepath.Join(tempWorkDir, "Procfile")
	if !common.FileExists(procfile) {
		return nil
	}

	b, err := sh.Command("procfile-util", "check", "-P", procfile).CombinedOutput()
	if err != nil {
		return fmt.Errorf(strings.TrimSpace(string(b[:])))
	}
	return nil
}

func TriggerPostStop(appName string) error {
	entries := map[string]string{
		"DOKKU_APP_RESTORE": "0",
	}

	return common.SuppressOutput(func() error {
		return config.SetMany(appName, entries, false)
	})
}

func TriggerPreDeploy(appName string, imageTag string) error {
	image := common.GetAppImageRepo(appName)
	removeProcfile(appName)

	procfilePath := getProcfilePath(appName)
	if err := extractProcfile(appName, image, procfilePath); err != nil {
		return err
	}
	if err := extractOrGenerateScalefile(appName, imageTag); err != nil {
		return err
	}

	return nil
}

func TriggerProcfileExtract(appName string, image string) error {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps", appName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	procfilePath := getProcfilePath(appName)

	if common.FileExists(procfilePath) {
		if err := common.PlugnTrigger("procfile-remove", []string{appName, procfilePath}...); err != nil {
			return err
		}
	}

	return extractProcfile(appName, image, procfilePath)
}

func TriggerProcfileGetCommand(appName string, processType string, port int) error {
	procfilePath := getProcfilePath(appName)
	if !common.FileExists(procfilePath) {
		image := common.GetDeployingAppImageName(appName, "", "")
		if err := common.PlugnTrigger("procfile-extract", []string{appName, image}...); err != nil {
			return err
		}
	}
	command, err := getProcfileCommand(procfilePath, processType, port)
	if err != nil {
		return err
	}

	if command != "" {
		fmt.Printf("%s\n", command)
	}

	return nil
}

func TriggerProcfileRemove(appName string, procfilePath string) error {
	if procfilePath == "" {
		procfilePath = getProcfilePath(appName)
	}

	if !common.FileExists(procfilePath) {
		return nil
	}

	os.Remove(procfilePath)
	return nil
}
