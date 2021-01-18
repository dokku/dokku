package cron

// TriggerPostDelete updates the cron entries for all apps
func TriggerPostDelete() error {
	return writeCronEntries()
}

// TriggerPostDeploy updates the cron entries for all apps
func TriggerPostDeploy() error {
	return writeCronEntries()
}
