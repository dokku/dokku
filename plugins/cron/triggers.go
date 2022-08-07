package cron

import (
	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall installs a sudoers file so we can execute crontab via sudo
func TriggerInstall() error {
	lines := []string{"%dokku ALL=(ALL) NOPASSWD:/usr/bin/crontab"}
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
