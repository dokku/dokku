package schedulerdockerlocal

// TriggerCronWrite force updates the cron file for all apps
func TriggerCronWrite(scheduler string) error {
	return writeCronEntries(scheduler)
}
