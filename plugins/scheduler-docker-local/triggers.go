package schedulerdockerlocal

// TriggerSchedulerCronWrite force updates the cron file for all apps
func TriggerSchedulerCronWrite(scheduler string) error {
	return writeCronTasks(scheduler)
}
