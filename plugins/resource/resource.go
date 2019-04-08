package resource

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// Resource is a collection of resource constraints for apps
type Resource struct {
	CPU            string `json:"cpu"`
	Memory         string `json:"memory"`
	MemorySwap     string `json:"memory-swap"`
	Network        string `json:"network"`
	NetworkIngress string `json:"network-ingress"`
	NetworkEgress  string `json:"network-egress"`
}

// ReportSingleApp is an internal function that displays the app report for one or more apps
func ReportSingleApp(appName, infoFlag string) {
	if err := common.VerifyAppName(appName); err != nil {
		common.LogFail(err.Error())
	}

	resources, err := common.PropertyGetAll("resource", appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	flags := []string{}
	infoFlags := map[string]string{}
	for key, value := range resources {
		flag := fmt.Sprintf("--resource-%v", key)
		flags = append(flags, flag)
		infoFlags[flag] = value
	}
	sort.Strings(flags)

	if len(infoFlag) == 0 {
		common.LogInfo2Quiet(fmt.Sprintf("%s resource information", appName))
		for _, k := range flags {
			v := infoFlags[k]
			key := strings.Replace(strings.Replace(strings.TrimPrefix(k, "--resource-"), "-", " ", -1), ".", " ", -1)
			common.LogVerbose(fmt.Sprintf("%s%s", Right(fmt.Sprintf("%s:", key), 31, " "), v))
		}
		return
	}

	for _, k := range flags {
		v := infoFlags[k]
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

// Fetches a single value for a given app/process/request/key combination
func GetResourceValue(appName string, processType string, requestType string, key string) (string, error) {
	resources, err := common.PropertyGetAll("resource", appName)
	if err != nil {
		return "", err
	}

	defaultValue := ""
	for key, value := range resources {
		if key == propertyKey("_default_", requestType, key) {
			defaultValue = value
		}
		if key == propertyKey(processType, requestType, key) {
			return value, nil
		}
	}

	return defaultValue, nil
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

func propertyKey(processType string, requestType string, key string) string {
	return fmt.Sprintf("%v.%v.%v", processType, requestType, key)
}

func getAppName(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("Please specify an app to run the command on")
	}

	appName := args[0]
	if err := common.VerifyAppName(appName); err != nil {
		return "", err
	}

	return appName, nil
}
