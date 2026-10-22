package resource

import (
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// validateResourceValue returns an error if the value is not valid for the given resource key
func validateResourceValue(key string, value string) error {
	switch key {
	case "memory":
		if !common.IsValidMemoryValue(value) {
			return fmt.Errorf("Invalid memory value '%s': must be a number with an optional unit suffix of b, k, m, g, t or p", value)
		}
	case "memory-swap":
		if value != "-1" && !common.IsValidMemoryValue(value) {
			return fmt.Errorf("Invalid memory-swap value '%s': must be -1 or a number with an optional unit suffix of b, k, m, g, t or p", value)
		}
	}

	return nil
}

// formatResourceValueForDocker validates a resource value and formats it for use as a docker flag value.
// Memory values without a unit default to megabytes
func formatResourceValueForDocker(key string, value string) (string, error) {
	if err := validateResourceValue(key, value); err != nil {
		return "", err
	}

	if key == "memory" && common.IsUnitlessMemoryValue(value) {
		value = value + "m"
	}

	return value, nil
}

// isResourceTypeKey returns true if the property key is a `<process>.<resource-type>.<resource>` key
// for the given resource type. The process type may contain dots, so the key is matched from the right
func isResourceTypeKey(key string, resourceType string) bool {
	parts := strings.Split(key, ".")
	if len(parts) < 3 {
		return false
	}

	return parts[len(parts)-2] == resourceType
}

func clearByResourceType(appName string, processType string, resourceType string) {
	noun := "limits"
	if resourceType == "reserve" {
		noun = "reservation"
	}

	message := fmt.Sprintf("clearing %v %v", appName, noun)
	if processType != "_default_" && processType != "" {
		message = fmt.Sprintf("%v (%v)", message, processType)
	}
	common.LogInfo2Quiet(message)

	if processType == "" {
		resources, err := common.PropertyGetAll("resource", appName)
		if err != nil {
			return
		}
		for key := range resources {
			if !isResourceTypeKey(key, resourceType) {
				continue
			}
			common.PropertyDelete("resource", appName, key)
		}
	} else {
		resources := []string{
			"cpu",
			"memory",
			"memory-swap",
			"network",
			"network-ingress",
			"network-egress",
			"nvidia-gpu",
		}

		for _, key := range resources {
			property := propertyKey(processType, resourceType, key)
			common.PropertyDelete("resource", appName, property)
		}
	}
}

func setResourceType(appName string, processType string, r Resource, resourceType string) error {
	if len(processType) == 0 {
		processType = "_default_"
	}

	resources := map[string]string{
		"cpu":             r.CPU,
		"memory":          r.Memory,
		"memory-swap":     r.MemorySwap,
		"network":         r.Network,
		"network-ingress": r.NetworkIngress,
		"network-egress":  r.NetworkEgress,
		"nvidia-gpu":      r.NvidiaGPU,
	}

	hasValues := false
	for _, value := range resources {
		if value != "" {
			hasValues = true
		}
	}

	if !hasValues {
		reportResourceType(appName, processType, resourceType)
		return nil
	}

	for key, value := range resources {
		if value == "" || value == "clear" {
			continue
		}

		if err := validateResourceValue(key, value); err != nil {
			return err
		}
	}

	noun := "limits"
	if resourceType == "reserve" {
		noun = "reservation"
	}
	message := fmt.Sprintf("Setting resource %v for %v", noun, appName)
	if processType != "_default_" {
		message = fmt.Sprintf("%v (%v)", message, processType)
	}
	common.LogInfo2Quiet(message)

	for key, value := range resources {
		if value == "" {
			continue
		}

		property := propertyKey(processType, resourceType, key)
		if value == "clear" {
			common.LogVerbose(fmt.Sprintf("%v: cleared", key))
			if err := common.PropertyDelete("resource", appName, property); err != nil {
				return err
			}
			continue
		}

		common.LogVerbose(fmt.Sprintf("%v: %v", key, value))
		if err := common.PropertyWrite("resource", appName, property, value); err != nil {
			return err
		}
	}

	return nil
}

func reportResourceType(appName string, processType string, resourceType string) {
	noun := "limits"
	if resourceType == "reserve" {
		noun = "reservation"
	}

	message := fmt.Sprintf("resource %v %v information", noun, appName)
	if processType == "_default_" {
		message = fmt.Sprintf("%v [defaults]", message)
	} else {
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
		"nvidia-gpu",
	}

	for _, key := range resources {
		property := propertyKey(processType, resourceType, key)
		value := common.PropertyGet("resource", appName, property)
		common.LogVerbose(fmt.Sprintf("%v: %v", key, value))
	}
}
