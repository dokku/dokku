package domains

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// backupDomainsTask is the body of a docket dokku_domains task.
type backupDomainsTask struct {
	// App is the name of the app. Empty when Global is true.
	App string `yaml:"app,omitempty"`

	// Global indicates the domains apply to the global scope.
	Global bool `yaml:"global,omitempty"`

	// State is the desired state of the domains.
	State string `yaml:"state"`

	// Domains is the list of domains to set.
	Domains []string `yaml:"domains"`
}

// backupDomainsToggleTask is the body of a docket dokku_domains_toggle task.
type backupDomainsToggleTask struct {
	// App is the name of the app.
	App string `yaml:"app"`

	// State is "absent" when vhost support is disabled for the app.
	State string `yaml:"state"`
}

// backupDomainsRecipe is the read shape for a domains backup slice.
type backupDomainsRecipe []struct {
	Tasks []struct {
		Toggle *backupDomainsToggleTask `yaml:"dokku_domains_toggle"`
	} `yaml:"tasks"`
}

// backupToggleDisabledState is the dokku_domains_toggle state for disabled vhosts.
const backupToggleDisabledState = "absent"

// backupVhostPath returns where a VHOST file lives within a backup scope dir.
func backupVhostPath(scopeDir string) string {
	return filepath.Join(common.BackupDataDir(scopeDir), "domains", "VHOST")
}

// TriggerBackupAppExport exports an app's domains and vhost toggle as a docket
// dokku_domains slice.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	tasks, err := exportVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "VHOST"), scopeDir, backupDomainsTask{App: appName})
	if err != nil {
		return err
	}

	if isAppVhostDisabled(appName) {
		tasks = append(tasks, common.BackupTask{"dokku_domains_toggle": backupDomainsToggleTask{
			App:   appName,
			State: backupToggleDisabledState,
		}})
	}

	return writeDomainsRecipe(scopeDir, tasks)
}

// TriggerBackupAppImport restores an app's domains and vhost toggle from the
// backup scope dir. The toggle always converges to the backup: an app without
// a disabled toggle in its slice, or without a slice, is left enabled.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	if err := importVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "VHOST"), scopeDir); err != nil {
		return err
	}

	disabled, err := backupVhostDisabled(scopeDir)
	if err != nil {
		return err
	}

	if disabled {
		return common.PropertyWrite("domains", appName, "disabled", "true")
	}

	return common.PropertyDelete("domains", appName, "disabled")
}

// TriggerBackupGlobalExport exports the global domains as a docket dokku_domains slice.
func TriggerBackupGlobalExport(scopeDir string) error {
	tasks, err := exportVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "VHOST"), scopeDir, backupDomainsTask{Global: true})
	if err != nil {
		return err
	}

	return writeDomainsRecipe(scopeDir, tasks)
}

// TriggerBackupGlobalImport restores the global domains from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return importVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "VHOST"), scopeDir)
}

// exportVhost copies a VHOST file into the backup scope dir and returns the
// matching dokku_domains task. An empty or missing file copies and returns nothing.
func exportVhost(vhostPath string, scopeDir string, task backupDomainsTask) ([]common.BackupTask, error) {
	contents := readFileString(vhostPath)
	for _, line := range strings.Split(contents, "\n") {
		if domain := strings.TrimSpace(line); domain != "" {
			task.Domains = append(task.Domains, domain)
		}
	}
	if len(task.Domains) == 0 {
		return nil, nil
	}

	destination := backupVhostPath(scopeDir)
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(destination, []byte(contents), 0600); err != nil {
		return nil, err
	}

	task.State = "set"
	return []common.BackupTask{{"dokku_domains": task}}, nil
}

// writeDomainsRecipe writes the domains recipe slice. No tasks writes nothing.
func writeDomainsRecipe(scopeDir string, tasks []common.BackupTask) error {
	if len(tasks) == 0 {
		return nil
	}

	recipe := common.BackupRecipe{{
		Name:  "domains",
		Tasks: tasks,
	}}
	return common.BackupWriteRecipe(common.BackupRecipePath("domains", scopeDir), recipe)
}

// backupVhostDisabled returns whether the domains slice in the backup scope
// dir disables vhost support. A missing slice is not disabled.
func backupVhostDisabled(scopeDir string) (bool, error) {
	path := common.BackupRecipePath("domains", scopeDir)
	if !common.FileExists(path) {
		return false, nil
	}

	var recipe backupDomainsRecipe
	if err := common.BackupReadRecipeFile(path, &recipe); err != nil {
		return false, err
	}

	for _, play := range recipe {
		for _, task := range play.Tasks {
			if task.Toggle != nil && task.Toggle.State == backupToggleDisabledState {
				return true, nil
			}
		}
	}

	return false, nil
}

// importVhost restores a VHOST file from the backup scope dir. A missing
// backup file is a no-op.
func importVhost(vhostPath string, scopeDir string) error {
	source := backupVhostPath(scopeDir)
	if !common.FileExists(source) {
		return nil
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return common.WriteBytesToFile(common.WriteBytesToFileInput{
		Bytes:    data,
		Filename: vhostPath,
		Mode:     0644,
	})
}
