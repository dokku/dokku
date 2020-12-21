package buildpacks

import (
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// CommandAdd implements buildpacks:add
func CommandAdd(appName string, buildpack string, index int) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	buildpack, err := validBuildpackURL(buildpack)
	if err != nil {
		return err
	}

	return common.PropertyListAdd("buildpacks", appName, "buildpacks", buildpack, index)
}

// CommandClear implements buildpacks:clear
func CommandClear(appName string) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
	}

	return common.PropertyDelete("buildpacks", appName, "buildpacks")
}

// CommandList implements buildpacks:list
func CommandList(appName string) (err error) {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
	if err != nil {
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
func CommandReport(appName string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}

// CommandSet implements buildpacks:set
func CommandSet(appName string, buildpack string, index int) error {
	if appName == "" {
		return errors.New("Please specify an app to run the command on")
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
