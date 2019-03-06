package buildpacks

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

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
