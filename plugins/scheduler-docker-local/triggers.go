package schedulerdockerlocal

// TriggerCronWrite force updates the cron file for all apps
func TriggerCronWrite() error {
	return writeCronEntries()
}
