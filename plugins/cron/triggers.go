package cron

import (
	"os"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall installs a sudoers file so we can execute crontab via sudo
func TriggerInstall() error {
	lines := []string{"%dokku ALL=(ALL) NOPASSWD:/usr/bin/crontab"}
	notty := map[string]bool{
		"centos": true,
		"fedora": true,
		"rhel":   true,
	}
	if notty[os.Getenv("DOKKU_DISTRO")] {
		lines = append(lines, "Defaults:dokku !requiretty")
	}

	return common.WriteSliceToFile("/etc/sudoers.d/dokku-cron", lines)
}

// TriggerPostDelete updates the cron entries for all apps
func TriggerPostDelete() error {
	return writeCronEntries()
}

// TriggerPostDeploy updates the cron entries for all apps
func TriggerPostDeploy() error {
	return writeCronEntries()
}

// TriggerCronWrite force updates the cron file for all apps
func TriggerCronWrite() error {
	return writeCronEntries()
}
