package resource

import (
	"errors"
	"fmt"

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
		return
	}

	infoFlags := map[string]string{}
	for key, value := range resources {
		flag := fmt.Sprintf("--resource-%v", key)
		infoFlags[flag] = value
	}

	trimPrefix := true
	uppercaseFirstCharacter := false
	common.ReportSingleApp("resource", appName, infoFlag, infoFlags, trimPrefix, uppercaseFirstCharacter)
}

// GetResourceValue fetches a single value for a given app/process/request/key combination
func GetResourceValue(appName string, processType string, resourceType string, key string) (string, error) {
	resources, err := common.PropertyGetAll("resource", appName)
	if err != nil {
		return "", err
	}

	defaultValue := ""
	for k, value := range resources {
		if k == propertyKey("_default_", resourceType, key) {
			defaultValue = value
		}
		if k == propertyKey(processType, resourceType, key) {
			return value, nil
		}
	}

	return defaultValue, nil
}

func propertyKey(processType string, resourceType string, key string) string {
	return fmt.Sprintf("%v.%v.%v", processType, resourceType, key)
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
