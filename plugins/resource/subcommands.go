package resource

import (
	"errors"
	"fmt"
	"github.com/dokku/dokku/plugins/common"
)

// CommandLimit implements resource:limit
func CommandLimit(args []string, processType string, r Resource) (err error) {
	return setRequestType(args, processType, r, "limit")
}

// CommandReserve implements resource:reserve
func CommandReserve(args []string, processType string, r Resource) (err error) {
	return setRequestType(args, processType, r, "reserve")
}

func setRequestType(args []string, processType string, r Resource, requestType string) (err error) {
	var appName string
	appName, err = getAppName(args)
	if err != nil {
		return
	}

	if err = common.VerifyAppName(appName); err != nil {
		common.LogFail(err.Error())
	}

	if len(processType) == 0 {
		processType = "_all_"
	}

	resources := map[string]string{
		"cpu":             r.Cpu,
		"memory":          r.Memory,
		"memory-swap":     r.MemorySwap,
		"network":         r.Network,
		"network-ingress": r.NetworkIngress,
		"network-egress":  r.NetworkEgress,
	}

	hasValues := false
	for _, value := range resources {
		if value != "" {
			hasValues = true
		}
	}

	if !hasValues {
		return errors.New("Please specify a resource to modify")
	}

	if requestType == "limit" {
		common.LogInfo2Quiet(fmt.Sprintf("Setting resource limits for %v", appName))
	} else if requestType == "reserve" {
		common.LogInfo2Quiet(fmt.Sprintf("Setting resource reservation for %v", appName))
	}

	for key, value := range resources {
		if value != "" {
			common.LogVerbose(fmt.Sprintf("%v: %v", key, value))
		}

		property := fmt.Sprintf("%v.%v.%v", processType, requestType, key)
		err = common.PropertyWrite(PluginName, appName, property, value)
		if err != nil {
			return
		}
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
