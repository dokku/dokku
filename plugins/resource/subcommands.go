package resource

import (
	"errors"
	"fmt"
	"github.com/dokku/dokku/plugins/common"
)

// CommandLimit implements resource:limit
func CommandLimit(args []string, processType string, r Resource, global bool) (err error) {
	appName, err = getAppName(args, global)
	if err != nil {
		return err
	}

	return setRequestType(appName, processType, r, "limit")
}

// CommandLimitClear implements resource:limit-clear
func CommandLimitClear(args []string, processType string, global bool) (err error) {
	appName, err = getAppName(args, global)
	if err != nil {
		return err
	}

	return clearByRequestType(appName, processType, "limit")
}

// CommandReserve implements resource:reserve
func CommandReserve(args []string, processType string, r Resource, global bool) (err error) {
	appName, err = getAppName(args, global)
	if err != nil {
		return err
	}

	return setRequestType(appName, processType, r, "reserve")
}

// CommandReserveClear implements resource:reserve-clear
func CommandReserveClear(args []string, processType string, global bool) error {
	appName, err = getAppName(args, global)
	if err != nil {
		return err
	}

	return clearByRequestType(appName, processType, "reserve")
}

func clearByRequestType(appName string, processType string, requestType string) error {
	noun := "limits"
	if requestType == "reserve" {
		noun = "reservation"
	}

	humanAppName := appName
	if appName == "_all_" {
		humanAppName = "default"
	}
	message := fmt.Sprintf("clearing %v %v", humanAppName, noun)
	if processType != "_all_" {
		message = fmt.Sprintf("%v (%v)", message, processType)
	}
	common.LogInfo2Quiet(message)

	resources := []string{
		"cpu",
		"memory",
		"memory-swap",
		"network",
		"network-ingress",
		"network-egress",
	}

	for _, key := range resources {
		property := propertyKey(processType, requestType, key)
		err = common.PropertyDelete("resource", appName, property)
		if err != nil {
			return err
		}
	}
	return nil
}

func setRequestType(appName string, processType string, r Resource, requestType string) (err error) {
	if len(processType) == 0 {
		processType = "_all_"
	}

	resources := map[string]string{
		"cpu":             r.CPU,
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
		return reportRequestType(appName, processType, requestType)
	}

	noun := "limits"
	if requestType == "reserve" {
		noun = "reservation"
	}
	message := fmt.Sprintf("Setting resource %v for %v", noun, appName)
	if appName == "_all_" {
		message = fmt.Sprintf("Setting default resource %v", noun)
	}

	if processType != "_all_" {
		message = fmt.Sprintf("%v (%v)", message, processType)
	}
	common.LogInfo2Quiet(message)

	for key, value := range resources {
		if value != "" {
			common.LogVerbose(fmt.Sprintf("%v: %v", key, value))
		}

		property := propertyKey(processType, requestType, key)
		err = common.PropertyWrite("resource", appName, property, value)
		if err != nil {
			return
		}
	}

	return
}

func reportRequestType(appName string, processType string, requestType string) (err error) {
	noun := "limits"
	if requestType == "reserve" {
		noun = "reservation"
	}

	humanAppName := appName
	if appName == "_all_" {
		humanAppName = "default"
	}
	message := fmt.Sprintf("resource %v %v information", noun, humanAppName)
	if processType != "_all_" {
		message = fmt.Sprintf("%v (%v)", message, processType)
	}
	common.LogInfo2Quiet(message)

	resources := []string{
		"cpu",
		"memory",
		"memory-swap",
		"network",
		"network-ingress",
		"network-egress",
	}

	for _, key := range resources {
		property := propertyKey(processType, requestType, key)
		value := common.PropertyGet("resource", appName, property)
		common.LogVerbose(fmt.Sprintf("%v: %v", key, value))
	}
	return nil
}

func propertyKey(processType string, requestType string, key string) string {
	return fmt.Sprintf("%v.%v.%v", processType, requestType, key)
}

func getAppName(args []string, global bool) (appName string, err error) {
	appName := "_all_"
	if global {
		return appName, nil
	}

	if len(args) < 1 {
		return "", errors.New("Please specify an app to run the command on")
	}

	appName = args[0]
	if err = common.VerifyAppName(appName); err != nil {
		return "", err
	}

	return appName, nil
}
