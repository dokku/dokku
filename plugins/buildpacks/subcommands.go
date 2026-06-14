package buildpacks

import (
	"errors"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// CommandAdd implements buildpacks:add
func CommandAdd(appName string, buildpack string, index int) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	buildpack, err := validBuildpackURL(buildpack)
	if err != nil {
		return err
	}

	return common.PropertyListAdd("buildpacks", appName, "buildpacks", buildpack, index)
}

// CommandClear implements buildpacks:clear
func CommandClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	return common.PropertyDelete("buildpacks", appName, "buildpacks")
}

// CommandDetect implements buildpacks:detect
func CommandDetect(appName string, branch string) (err error) {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	workDir := common.AppRoot(appName)
	checkedOutDir, err := checkoutBareGitRepo(workDir, branch)
    if err != nil {
        return err
    }
	defer func() {
		if err := os.RemoveAll(checkedOutDir); err != nil {
			common.LogWarn(fmt.Sprintf("Failed to remove temporary directory %s: %v", checkedOutDir, err))
		}
	}()

	dockerArgs := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/tmp/app", checkedOutDir),
		"gliderlabs/herokuish", "/bin/herokuish", "buildpack", "detect", "/tmp/app",
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    dockerArgs,
	})
	if err != nil {
		return fmt.Errorf("Buildpack detection failed: %s", result.StderrContents())
	}

	common.LogVerbose(result.StdoutContents())
	return nil
}

// CommandList implements buildpacks:list
func CommandList(appName string) (err error) {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return
	}

	common.LogInfo1Quiet(fmt.Sprintf("%s buildpack urls", appName))
	for _, buildpack := range buildpacks {
		common.LogVerbose(buildpack)
	}
	return nil
}

// CommandRemove implements buildpacks:remove
func CommandRemove(appName string, buildpack string, index int) (err error) {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if index != 0 && buildpack != "" {
		err = errors.New("Please choose either index or Buildpack, but not both")
		return
	}

	if index == 0 && buildpack == "" {
		err = errors.New("Must specify a buildpack to remove, either by index or URL")
		return
	}

	buildpack, err = validBuildpackURL(buildpack)
	if index == 0 && err != nil {
		return err
	}

	var buildpacks []string
	buildpacks, err = common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return
	}

	if len(buildpacks) == 0 {
		err = fmt.Errorf("No buildpacks were found, next release on %s will detect buildpack normally", appName)
		return
	}

	if index != 0 {
		var value string
		value, err = common.PropertyListGetByIndex("buildpacks", appName, "buildpacks", index-1)
		if err != nil {
			return
		}

		buildpack = value
	} else {
		_, err = common.PropertyListGetByValue("buildpacks", appName, "buildpacks", buildpack)
		if err != nil {
			return
		}
	}

	common.LogInfo1Quiet(fmt.Sprintf("Removing %s", buildpack))
	err = common.PropertyListRemove("buildpacks", appName, "buildpacks", buildpack)
	if err != nil {
		return
	}
	return
}

// CommandReport displays a buildpacks report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandSet implements buildpacks:set
func CommandSet(appName string, buildpack string, index int) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	buildpack, err := validBuildpackURL(buildpack)
	if err != nil {
		return err
	}

	if index > 0 {
		index--
	}

	return common.PropertyListSet("buildpacks", appName, "buildpacks", buildpack, index)
}

// CommandSetProperty implements buildpacks:set-property
func CommandSetProperty(appName string, property string, value string) error {
	if property != "stack" {
		common.CommandPropertySet("buildpacks", appName, property, value, DefaultProperties, GlobalProperties)
		return nil
	}

	if value != "" {
		builder := builderForStack(value)
		common.LogWarn(fmt.Sprintf("Deprecated: buildpacks:set-property stack is deprecated, use %s:set stack instead", builder))
		if err := common.PropertyWrite(builder, appName, "stack", value); err != nil {
			return err
		}
		return clearBuilderCache(builder, appName)
	}

	common.LogWarn("Deprecated: buildpacks:set-property stack is deprecated, use builder-herokuish:set stack or builder-pack:set stack instead")
	for _, builder := range []string{"builder-herokuish", "builder-pack"} {
		if err := common.PropertyDelete(builder, appName, "stack"); err != nil {
			return err
		}
		if err := clearBuilderCache(builder, appName); err != nil {
			return err
		}
	}

	return nil
}

// clearBuilderCache clears the build cache for the given builder after a stack
// change. The herokuish cache is the cache-$APP volume removed by repo via the
// post-stack-set trigger, while the pack cache lives in pack-managed volumes and
// is cleared by flagging the next build to run with --clear-cache.
func clearBuilderCache(builder string, appName string) error {
	apps := []string{appName}
	if appName == "--global" {
		dokkuApps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				return nil
			}
			return err
		}
		apps = dokkuApps
	}

	for _, app := range apps {
		if builder == "builder-herokuish" {
			if _, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
				Trigger:     "post-stack-set",
				Args:        []string{app, ""},
				StreamStdio: true,
			}); err != nil {
				return err
			}
			continue
		}

		if err := common.PropertyWrite("builder-pack", app, "clear-cache", "true"); err != nil {
			return err
		}
	}

	return nil
}
