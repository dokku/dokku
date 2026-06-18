package backup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// CommandExport implements `dokku backup:export`. It resolves the requested
// scope, dispatches the export triggers to every plugin, writes the manifest
// and combined recipes, and produces a single tar.gz. The absolute path of the
// finished tarball is the only thing written to stdout; all logging goes to
// stderr so callers can capture the path.
func CommandExport(appNames []string, serviceSpecs []string, backupDir string, includeStorage bool) error {
	createdAt := time.Now().UTC()
	if backupDir == "" {
		backupDir = os.TempDir()
	}
	if err := ensureBackupDirWritable(backupDir); err != nil {
		return err
	}

	apps, services, scopeLabel, err := resolveExportScope(appNames, serviceSpecs)
	if err != nil {
		return err
	}

	stagingDir, err := os.MkdirTemp(backupDir, ".dokku-backup-staging-*")
	if err != nil {
		return fmt.Errorf("unable to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	timestamp := createdAt.Format("20060102T150405Z")
	finalPath := filepath.Join(backupDir, fmt.Sprintf("dokku-backup-%s-%s.tar.gz", scopeLabel, timestamp))

	tmpTar, err := os.CreateTemp(backupDir, ".dokku-backup-*.tar.gz.tmp")
	if err != nil {
		return fmt.Errorf("unable to create temporary tarball: %w", err)
	}
	tmpTarPath := tmpTar.Name()
	tmpTar.Close()
	defer os.Remove(tmpTarPath)

	manifest := NewManifest(createdAt.Format(time.RFC3339), includeStorage)

	common.LogStderr(fmt.Sprintf("Exporting backup (%d app(s), %d service(s))", len(apps), len(services)))

	if err := dispatchTrigger("pre-backup-export", stagingDir, finalPath); err != nil {
		return err
	}

	globalDir := globalScopeDir(stagingDir)
	if err := makeScopeDirs(globalDir); err != nil {
		return err
	}
	if err := dispatchTrigger("backup-global-export", globalDir); err != nil {
		return err
	}

	storageEnv := map[string]string{"DOKKU_QUIET_OUTPUT": "1"}
	if includeStorage {
		storageEnv["DOKKU_BACKUP_INCLUDE_STORAGE"] = "1"
	}

	for _, appName := range apps {
		scope := appScopeDir(stagingDir, appName)
		if err := makeScopeDirs(scope); err != nil {
			return err
		}
		if err := dispatchTrigger("pre-backup-app-export", appName, scope); err != nil {
			return err
		}
		if err := dispatchTriggerWithEnv(storageEnv, "backup-app-export", appName, scope); err != nil {
			return err
		}
		if err := dispatchTrigger("post-backup-app-export", appName, scope); err != nil {
			return err
		}
		manifest.Apps[appName] = collectAppMeta(appName, scope)
	}

	for _, ref := range services {
		scope := serviceScopeDir(stagingDir, ref.Type, ref.Name)
		if err := makeScopeDirs(scope); err != nil {
			return err
		}
		if err := dispatchTrigger("pre-backup-service-export", ref.Type, ref.Name, scope); err != nil {
			return err
		}
		if err := dispatchTrigger("backup-service-export", ref.Type, ref.Name, scope); err != nil {
			return err
		}
		if err := dispatchTrigger("post-backup-service-export", ref.Type, ref.Name, scope); err != nil {
			return err
		}
		manifest.Services[ref.key()] = ManifestService{}
	}

	if err := WriteManifest(filepath.Join(stagingDir, ManifestFilename), manifest); err != nil {
		return err
	}

	if err := assembleRecipes(stagingDir, apps, services); err != nil {
		common.LogWarn(fmt.Sprintf("unable to assemble combined recipe: %v", err))
	}

	if err := setOwnership(stagingDir); err != nil {
		common.LogWarn(fmt.Sprintf("unable to set ownership on staging dir: %v", err))
	}

	if err := TarGzCreate(stagingDir, tmpTarPath); err != nil {
		return err
	}
	if err := os.Rename(tmpTarPath, finalPath); err != nil {
		return fmt.Errorf("unable to finalize backup file: %w", err)
	}
	if err := common.SetPermissions(common.SetPermissionInput{Filename: finalPath, Mode: 0600}); err != nil {
		common.LogWarn(fmt.Sprintf("unable to set ownership on backup file: %v", err))
	}

	if err := dispatchTrigger("post-backup-export", stagingDir, finalPath); err != nil {
		return err
	}

	fmt.Println(finalPath)
	return nil
}

// resolveExportScope determines the apps and services to export and a label for
// the filename. With no filters everything is exported; explicit filters are
// validated and access-checked.
func resolveExportScope(appNames []string, serviceSpecs []string) ([]string, []serviceRef, string, error) {
	explicitApp := len(appNames) > 0
	explicitService := len(serviceSpecs) > 0

	if !explicitApp && !explicitService {
		apps, err := common.DokkuApps()
		if err != nil && !errors.Is(err, common.NoAppsExist) {
			return nil, nil, "", err
		}
		services, err := discoverAllServices()
		if err != nil {
			return nil, nil, "", err
		}
		return apps, services, "full", nil
	}

	var apps []string
	for _, appName := range appNames {
		if err := common.VerifyAppName(appName); err != nil {
			return nil, nil, "", err
		}
		if !callerHasAppAccess(appName) {
			return nil, nil, "", fmt.Errorf("Unable to back up %s", appName)
		}
		apps = append(apps, appName)
	}

	var services []serviceRef
	for _, spec := range serviceSpecs {
		serviceType, serviceName := parseServiceSpec(spec)
		if serviceName == "" {
			refs, err := listServicesOfType(serviceType)
			if err != nil {
				return nil, nil, "", err
			}
			for _, ref := range refs {
				if callerHasServiceAccess(ref) {
					services = append(services, ref)
				}
			}
			continue
		}
		ref := serviceRef{Type: serviceType, Name: serviceName}
		if !callerHasServiceAccess(ref) {
			return nil, nil, "", fmt.Errorf("Unable to back up %s", ref)
		}
		services = append(services, ref)
	}

	return apps, services, exportLabel(apps, services), nil
}

// exportLabel returns a short filename label for an explicit-scope export.
func exportLabel(apps []string, services []serviceRef) string {
	if len(apps) == 1 && len(services) == 0 {
		return apps[0]
	}
	if len(apps) == 0 && len(services) == 1 {
		return fmt.Sprintf("%s-%s", services[0].Type, services[0].Name)
	}
	return "partial"
}

// makeScopeDirs creates the config/ and data/ sub-trees of a scope directory.
func makeScopeDirs(scopeDir string) error {
	if err := os.MkdirAll(common.BackupConfigDir(scopeDir), 0755); err != nil {
		return err
	}
	return os.MkdirAll(common.BackupDataDir(scopeDir), 0755)
}

// collectAppMeta derives per-app manifest metadata from the produced scope dir.
func collectAppMeta(appName string, scopeDir string) ManifestApp {
	meta := ManifestApp{
		Deployed:  common.IsDeployed(appName),
		Scheduler: common.GetAppScheduler(appName),
	}

	if common.FileExists(filepath.Join(common.BackupDataDir(scopeDir), "repo.bundle")) {
		meta.HasCode = true
	}

	configDir := common.BackupConfigDir(scopeDir)
	if entries, err := os.ReadDir(configDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if filepath.Ext(name) == ".yml" && name != "tasks.yml" {
				meta.Plugins = append(meta.Plugins, name[:len(name)-len(".yml")])
			}
		}
	}

	dataDir := common.BackupDataDir(scopeDir)
	filepath.Walk(dataDir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if rel, relErr := filepath.Rel(scopeDir, p); relErr == nil {
			meta.DataFiles = append(meta.DataFiles, filepath.ToSlash(rel))
		}
		return nil
	})

	return meta
}
