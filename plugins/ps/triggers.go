package ps

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	sh "github.com/codeskyblue/go-sh"
	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	dockeroptions "github.com/dokku/dokku/plugins/docker-options"
	"github.com/otiai10/copy"
)

// TriggerAppRestart restarts an app
func TriggerAppRestart(appName string) error {
	return Restart(appName)
}

// TriggerCorePostDeploy sets a property to
// allow the app to be restored on boot
func TriggerCorePostDeploy(appName string) error {
	existingProcfile := getProcfilePath(appName)
	processSpecificProcfile := fmt.Sprintf("%s.%s", existingProcfile, os.Getenv("DOKKU_PID"))
	if common.FileExists(processSpecificProcfile) {
		if err := os.Rename(processSpecificProcfile, existingProcfile); err != nil {
			return err
		}
	} else if common.FileExists(fmt.Sprintf("%s.missing", processSpecificProcfile)) {
		if err := os.Remove(fmt.Sprintf("%s.missing", processSpecificProcfile)); err != nil {
			return err
		}

		if err := os.Remove(existingProcfile); err != nil {
			return err
		}
	}

	entries := map[string]string{
		"DOKKU_APP_RESTORE": "1",
	}

	return common.SuppressOutput(func() error {
		return config.SetMany(appName, entries, false)
	})
}

// TriggerCorePostExtract ensures that the main Procfile is the one specified by procfile-path
func TriggerCorePostExtract(appName string, sourceWorkDir string) error {
	procfilePath := strings.Trim(reportComputedProcfilePath(appName), "/")
	if procfilePath == "" {
		procfilePath = "Procfile"
	}

	existingProcfile := getProcfilePath(appName)

	files, err := filepath.Glob(fmt.Sprintf("%s.*", existingProcfile))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	repoProcfilePath := path.Join(sourceWorkDir, procfilePath)
	processSpecificProcfile := fmt.Sprintf("%s.%s", existingProcfile, os.Getenv("DOKKU_PID"))
	if !common.FileExists(repoProcfilePath) {
		return common.TouchFile(fmt.Sprintf("%s.missing", processSpecificProcfile))
	}

	if err := copy.Copy(repoProcfilePath, processSpecificProcfile); err != nil {
		return fmt.Errorf("Unable to extract Procfile: %v", err.Error())
	}

	b, err := sh.Command("procfile-util", "check", "-P", processSpecificProcfile).CombinedOutput()
	if err != nil {
		return fmt.Errorf(strings.TrimSpace(string(b[:])))
	}
	return nil
}

// TriggerInstall initializes app restart policies
func TriggerInstall() error {
	if err := common.PropertySetup("ps"); err != nil {
		return fmt.Errorf("Unable to install the ps plugin: %s", err.Error())
	}

	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	if err := common.SetPermissions(directory, 0755); err != nil {
		return err
	}

	apps, err := common.UnfilteredDokkuApps()
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

	for _, appName := range apps {
		dokkuScaleFile := filepath.Join(common.AppRoot(appName), "DOKKU_SCALE")
		if common.FileExists(dokkuScaleFile) {
			processTuples, err := common.FileToSlice(dokkuScaleFile)
			if err != nil {
				return err
			}

			if err := scaleSet(appName, true, false, processTuples); err != nil {
				return err
			}

			os.Remove(dokkuScaleFile)
		}

		dokkuScaleExtracted := filepath.Join(common.AppRoot(appName), "DOKKU_SCALE.extracted")
		if common.FileExists(dokkuScaleExtracted) {
			if err := common.PropertyWrite("ps", appName, "can-scale", strconv.FormatBool(false)); err != nil {
				return err
			}
			os.Remove(dokkuScaleExtracted)
		}
	}

	return nil
}

// TriggerPostAppClone rebuilds the new app
func TriggerPostAppClone(oldAppName string, newAppName string) error {
	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

// TriggerPostAppCloneSetup creates new ps files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("ps", oldAppName, newAppName)
	if err != nil {
		return err
	}

	// TODO: Copy data dir
	return nil
}

// TriggerPostAppRename rebuilds the renamed app
func TriggerPostAppRename(oldAppName string, newAppName string) error {
	if os.Getenv("SKIP_REBUILD") == "true" {
		return nil
	}

	return Rebuild(newAppName)
}

// TriggerPostAppRenameSetup renames ps files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("ps", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("ps", oldAppName); err != nil {
		return err
	}

	// TODO: Move data dir
	return nil
}

// TriggerPostCreate ensures apps have a default restart policy
// and scale value for web
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

	formations := FormationSlice{
		&Formation{
			ProcessType: "web",
			Quantity:    1,
		},
	}
	return updateScale(appName, false, formations)
}

// TriggerPostDelete destroys the ps properties for a given app container
func TriggerPostDelete(appName string) error {
	directory := filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "data", "ps", appName)
	dataErr := os.RemoveAll(directory)
	propertyErr := common.PropertyDestroy("ps", appName)

	if dataErr != nil {
		return dataErr
	}

	return propertyErr
}

// TriggerPostStop sets the restore property to false
func TriggerPostStop(appName string) error {
	entries := map[string]string{
		"DOKKU_APP_RESTORE": "0",
	}

	return common.SuppressOutput(func() error {
		return config.SetMany(appName, entries, false)
	})
}

// TriggerPreDeploy ensures an app has an up to date scale parameters
func TriggerPreDeploy(appName string, imageTag string) error {
	if err := updateScale(appName, false, FormationSlice{}); err != nil {
		common.LogDebug(fmt.Sprintf("Error generating scale file: %s", err.Error()))
		return err
	}

	return nil
}

// TriggerProcfileGetCommand fetches a command from the procfile
func TriggerProcfileGetCommand(appName string, processType string, port int) error {
	procfilePath := getProcfilePath(appName)
	command, err := getProcfileCommand(procfilePath, processType, port)
	if err != nil {
		return err
	}

	if command != "" {
		fmt.Printf("%s\n", command)
	}

	return nil
}

// TriggerPsCanScale sets whether or not a user can scale an app with ps:scale
func TriggerPsCanScale(appName string, canScale bool) error {
	return common.PropertyWrite("ps", appName, "can-scale", strconv.FormatBool(canScale))
}

// TriggerPsCurrentScale prints out the current scale contents (process-type=quantity) delimited by newlines
func TriggerPsCurrentScale(appName string) error {
	formations, err := getFormations(appName)
	if err != nil {
		return err
	}

	sort.Sort(formations)
	lines := []string{}
	for _, formation := range formations {
		lines = append(lines, fmt.Sprintf("%s=%d", formation.ProcessType, formation.Quantity))
	}

	fmt.Print(strings.Join(lines, "\n"))

	return nil
}

// TriggerPsSetScale configures the scale parameters for a given app
func TriggerPsSetScale(appName string, skipDeploy bool, clearExisting bool, processTuples []string) error {
	return scaleSet(appName, skipDeploy, clearExisting, processTuples)
}
