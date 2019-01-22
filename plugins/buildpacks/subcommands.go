package buildpacks

import (
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// CommandAdd implements buildpacks:add
func CommandAdd(args []string, index int) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
	}

	buildpack := ""
	if len(args) >= 2 {
		buildpack = args[1]
	}
	if buildpack == "" {
		return errors.New("Must specify a buildpack to add")
	}

	err = common.PropertyListAdd("buildpacks", appName, "buildpacks", buildpack, index)
	return
}

// CommandClear implements buildpacks:clear
func CommandClear(args []string) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
	}

	return common.PropertyDelete("buildpacks", appName, "buildpacks")
}

// CommandList implements buildpacks:list
func CommandList(args []string) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
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
func CommandRemove(args []string, index int) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
	}

	buildpack := ""
	if len(args) >= 2 {
		buildpack = args[1]
	}
	if index != 0 && buildpack != "" {
		err = errors.New("Please choose either index or Buildpack, but not both")
		return
	}

	if index == 0 && buildpack == "" {
		err = errors.New("Must specify a buildpack to remove, either by index or URL")
		return
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

// CommandSet implements buildpacks:set
func CommandSet(args []string, index int) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
	}

	buildpack := ""
	if len(args) >= 2 {
		buildpack = args[1]
	}
	if buildpack == "" {
		return errors.New("Must specify a buildpack to add")
	}

	if index > 0 {
		index--
	}

	err = common.PropertyListSet("buildpacks", appName, "buildpacks", buildpack, index)
	if err != nil {
		return
	}

	return
}

func getAppName(args []string) (appName string, err error) {
	if len(args) >= 1 {
		appName = args[0]
	} else {
		err = errors.New("Please specify an app to run the command on")
	}

	return
}
