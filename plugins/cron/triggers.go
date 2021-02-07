package cron

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
