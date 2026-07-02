package backup

// TriggerInstall runs the install step for the backup plugin. The plugin holds
// no persistent state of its own, so installation is a no-op today.
func TriggerInstall() error {
	return nil
}
