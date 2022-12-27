package proxy

import (
	"fmt"
)

// TriggerProxyIsEnabled prints true or false depending on whether the proxy is enabled
func TriggerProxyIsEnabled(appName string) error {
	if IsAppProxyEnabled(appName) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}

	return nil
}

// TriggerProxyType prints out the current proxy type, defaulting to nginx
func TriggerProxyType(appName string) error {
	proxyType := getAppProxyType(appName)
	fmt.Println(proxyType)

	return nil
}
