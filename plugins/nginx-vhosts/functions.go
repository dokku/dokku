package nginxvhosts

import (
	"os"
)

func getLogRoot() string {
	logRoot := "/var/log/nginx"
	if isOpenRestyInstalled() {
		logRoot = "/var/log/openresty"
	}
	return logRoot
}

func isOpenRestyInstalled() bool {
	fi, err := os.Stat("/usr/bin/openresty")
	if err != nil {
		return false
	}

	if fi.IsDir() {
		return false
	}

	return true
}
