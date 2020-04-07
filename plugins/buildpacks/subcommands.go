package buildpacks

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func validBuildpackURL(buildpack string) (string, error) {
	if buildpack == "" {
		return buildpack, errors.New("Must specify a buildpack to add")
	}

	reHerokuValue := regexp.MustCompile(`(?m)^([\w]+\/[\w]+)$`)
	if found := reHerokuValue.Find([]byte(buildpack)); found != nil {
		parts := strings.SplitN(buildpack, "/", 2)
		return fmt.Sprintf("https://github.com/%s/heroku-buildpack-%s.git", parts[0], parts[1]), nil
	}

	reString := regexp.MustCompile(`(?m)^(http|https|git)(:\/\/|@)([^\/:]+)[\/:]([^\/:]+)\/(.+)(.git(#derp)?)?$`)
	if found := reString.Find([]byte(buildpack)); found != nil {
		return buildpack, nil
	}

	return buildpack, fmt.Errorf("Invalid buildpack specified: %v", buildpack)
}

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

	buildpack, err = validBuildpackURL(buildpack)
	if err != nil {
		return err
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

	common.PropertyDelete("buildpacks", appName, "buildpacks")
	return
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
	if strings.HasPrefix(appName, "--") {
		infoFlag = appName
		appName = ""
	}

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

	buildpack, err = validBuildpackURL(buildpack)
	if err != nil {
		return err
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
