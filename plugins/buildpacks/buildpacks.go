package buildpacks

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// PostExtract is a plugin trigger that writes a .buildpacks file into the app
func PostExtract(appName, tmpWorkDir string) {
	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return
	}

	if len(buildpacks) == 0 {
		return
	}

	buildpacksPath := path.Join(tmpWorkDir, ".buildpacks")
	file, err := os.OpenFile(buildpacksPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}

	w := bufio.NewWriter(file)
	for _, buildpack := range buildpacks {
		fmt.Fprintln(w, buildpack)
	}

	if err = w.Flush(); err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}
	file.Chmod(0600)
}

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) {
	if err := common.VerifyAppName(appName); err != nil {
		common.LogFail(err.Error())
	}

	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		common.LogFail(err.Error())
	}

	infoFlags := map[string]string{
		"--buildpacks-list": strings.Join(buildpacks, ","),
	}

	if len(infoFlag) == 0 {
		common.LogInfo2Quiet(fmt.Sprintf("%s buildpacks information", appName))
		for k, v := range infoFlags {
			key := common.UcFirst(strings.Replace(strings.TrimPrefix(k, "--"), "-", " ", -1))
			common.LogVerbose(fmt.Sprintf("%s%s", Right(fmt.Sprintf("%s:", key), 31, " "), v))
		}
		return
	}

	for k, v := range infoFlags {
		if infoFlag == k {
			fmt.Fprintln(os.Stdout, v)
			return
		}
	}

	keys := reflect.ValueOf(infoFlags).MapKeys()
	strkeys := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		strkeys[i] = keys[i].String()
	}
	common.LogFail(fmt.Sprintf("Invalid flag passed, valid flags: %s", strings.Join(strkeys, ", ")))
}

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Right right-pads the string with pad up to len runes
func Right(str string, length int, pad string) string {
	return str + times(pad, length-len(str))
}
