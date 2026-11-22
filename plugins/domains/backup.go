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

// backupVhostPath returns where a VHOST file lives within a backup scope dir.
func backupVhostPath(scopeDir string) string {
	return filepath.Join(common.BackupDataDir(scopeDir), "domains", "VHOST")
}

// TriggerBackupAppExport exports an app's domains as a docket dokku_domains slice.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	return exportVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "VHOST"), scopeDir, backupDomainsTask{App: appName})
}

// TriggerBackupAppImport restores an app's domains from the backup scope dir.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	return importVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "VHOST"), scopeDir)
}

// TriggerBackupGlobalExport exports the global domains as a docket dokku_domains slice.
func TriggerBackupGlobalExport(scopeDir string) error {
	return exportVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "VHOST"), scopeDir, backupDomainsTask{Global: true})
}

// TriggerBackupGlobalImport restores the global domains from the backup scope dir.
func TriggerBackupGlobalImport(scopeDir string) error {
	return importVhost(filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "VHOST"), scopeDir)
}

// exportVhost copies a VHOST file into the backup scope dir and writes the
// matching dokku_domains recipe slice. An empty or missing file writes nothing.
func exportVhost(vhostPath string, scopeDir string, task backupDomainsTask) error {
	contents := readFileString(vhostPath)
	for _, line := range strings.Split(contents, "\n") {
		if domain := strings.TrimSpace(line); domain != "" {
			task.Domains = append(task.Domains, domain)
		}
	}
	if len(task.Domains) == 0 {
		return nil
	}

	destination := backupVhostPath(scopeDir)
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(destination, []byte(contents), 0600); err != nil {
		return err
	}

	task.State = "set"
	recipe := common.BackupRecipe{{
		Name:  "domains",
		Tasks: []common.BackupTask{{"dokku_domains": task}},
	}}
	return common.BackupWriteRecipe(common.BackupRecipePath("domains", scopeDir), recipe)
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
