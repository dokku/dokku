package domains

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
	. "github.com/onsi/gomega"
)

func TestBackupAppRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t, "app-a")
	vhostPath := filepath.Join(dokkuRoot, "app-a", "VHOST")
	writeVhost(vhostPath, "a.example.com\nb.example.com\n")

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupAppExport("app-a", scope)).To(Succeed())

	recipe, err := os.ReadFile(common.BackupRecipePath("domains", scope))
	Expect(err).To(Succeed())
	Expect(string(recipe)).To(ContainSubstring("dokku_domains:"))
	Expect(string(recipe)).NotTo(ContainSubstring("dokku_domains_toggle"))

	Expect(os.Remove(vhostPath)).To(Succeed())
	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())

	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(readFileString(vhostPath)).To(Equal("a.example.com\nb.example.com\n"))
	Expect(common.PropertyExists("domains", "app-a", "disabled")).To(BeFalse())
}

func TestBackupAppDisabledRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t, "app-a")
	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupAppExport("app-a", scope)).To(Succeed())
	Expect(common.FileExists(backupVhostPath(scope))).To(BeFalse())

	recipe, err := os.ReadFile(common.BackupRecipePath("domains", scope))
	Expect(err).To(Succeed())
	Expect(string(recipe)).To(ContainSubstring("dokku_domains_toggle:"))
	Expect(string(recipe)).To(ContainSubstring("app: app-a"))
	Expect(string(recipe)).To(ContainSubstring("state: absent"))
	Expect(string(recipe)).NotTo(ContainSubstring("dokku_domains:"))

	Expect(common.PropertyDelete("domains", "app-a", "disabled")).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(common.PropertyGet("domains", "app-a", "disabled")).To(Equal("true"))
	Expect(common.FileExists(filepath.Join(dokkuRoot, "app-a", "VHOST"))).To(BeFalse())

	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(common.PropertyGet("domains", "app-a", "disabled")).To(Equal("true"))
}

func TestBackupAppDisabledWithDomains(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t, "app-a")
	writeVhost(filepath.Join(dokkuRoot, "app-a", "VHOST"), "a.example.com\n")
	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupAppExport("app-a", scope)).To(Succeed())

	var recipe common.BackupRecipe
	Expect(common.BackupReadRecipeFile(common.BackupRecipePath("domains", scope), &recipe)).To(Succeed())
	Expect(recipe).To(HaveLen(1))
	Expect(recipe[0].Tasks).To(HaveLen(2))
	Expect(recipe[0].Tasks[0]).To(HaveKey("dokku_domains"))
	Expect(recipe[0].Tasks[1]).To(HaveKey("dokku_domains_toggle"))

	Expect(common.PropertyDestroy("domains", "app-a")).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(common.PropertyGet("domains", "app-a", "disabled")).To(Equal("true"))
}

func TestBackupAppWithoutSlice(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t, "app-a")

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupAppExport("app-a", scope)).To(Succeed())
	Expect(common.FileExists(common.BackupRecipePath("domains", scope))).To(BeFalse())

	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(common.PropertyExists("domains", "app-a", "disabled")).To(BeFalse())
}

func TestBackupAppToggleEnabledState(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t, "app-a")

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(writeDomainsRecipe(scope, []common.BackupTask{{"dokku_domains_toggle": backupDomainsToggleTask{App: "app-a", State: "present"}}})).To(Succeed())

	Expect(common.PropertyWrite("domains", "app-a", "disabled", "true")).To(Succeed())
	Expect(TriggerBackupAppImport("app-a", scope)).To(Succeed())
	Expect(common.PropertyExists("domains", "app-a", "disabled")).To(BeFalse())
}

func TestBackupAppInvalidSlice(t *testing.T) {
	RegisterTestingT(t)
	_, libRoot := setupTestEnv(t, "app-a")

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	path := common.BackupRecipePath("domains", scope)
	Expect(os.MkdirAll(filepath.Dir(path), 0755)).To(Succeed())
	Expect(os.WriteFile(path, []byte("not: [valid"), 0600)).To(Succeed())

	Expect(TriggerBackupAppImport("app-a", scope)).To(HaveOccurred())
}

func TestBackupAppExportErrors(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t, "app-a")
	writeVhost(filepath.Join(dokkuRoot, "app-a", "VHOST"), "a.example.com\n")

	blocker := filepath.Join(libRoot, "blocker")
	writeVhost(blocker, "")
	Expect(TriggerBackupAppExport("app-a", blocker)).To(HaveOccurred())
	Expect(TriggerBackupGlobalExport(blocker)).To(Succeed())

	writeVhost(filepath.Join(dokkuRoot, "VHOST"), "dokku.me\n")
	Expect(TriggerBackupGlobalExport(blocker)).To(HaveOccurred())

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(os.MkdirAll(backupVhostPath(scope), 0755)).To(Succeed())
	Expect(TriggerBackupAppExport("app-a", scope)).To(HaveOccurred())
}

func TestBackupAppImportErrors(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t, "app-a")
	writeVhost(filepath.Join(dokkuRoot, "app-a", "VHOST"), "a.example.com\n")

	scope := filepath.Join(libRoot, "backup", "apps", "app-a")
	Expect(TriggerBackupAppExport("app-a", scope)).To(Succeed())
	Expect(TriggerBackupAppImport("missing-app", scope)).To(HaveOccurred())
}

func TestBackupGlobalRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	dokkuRoot, libRoot := setupTestEnv(t)
	vhostPath := filepath.Join(dokkuRoot, "VHOST")

	scope := filepath.Join(libRoot, "backup", "global")
	Expect(TriggerBackupGlobalExport(scope)).To(Succeed())
	Expect(common.FileExists(common.BackupRecipePath("domains", scope))).To(BeFalse())
	Expect(TriggerBackupGlobalImport(scope)).To(Succeed())
	Expect(common.FileExists(vhostPath)).To(BeFalse())

	writeVhost(vhostPath, "dokku.me\n")
	Expect(TriggerBackupGlobalExport(scope)).To(Succeed())

	recipe, err := os.ReadFile(common.BackupRecipePath("domains", scope))
	Expect(err).To(Succeed())
	Expect(string(recipe)).To(ContainSubstring("global: true"))
	Expect(string(recipe)).NotTo(ContainSubstring("dokku_domains_toggle"))

	Expect(os.Remove(vhostPath)).To(Succeed())
	Expect(TriggerBackupGlobalImport(scope)).To(Succeed())
	Expect(readFileString(vhostPath)).To(Equal("dokku.me\n"))
}
