package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandImport implements `dokku backup:import`. It extracts the backup,
// validates the manifest, enforces per-item access control, then restores the
// global scope, services, and apps in order. Restore is best-effort per item:
// a failed or access-denied item is skipped and reported in the closing
// summary; the rest of the import proceeds.
func CommandImport(backupFile string, appNames []string, serviceSpecs []string, force bool, installPlugins bool) error {
	if !common.FileExists(backupFile) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	workDir, err := os.MkdirTemp(os.TempDir(), ".dokku-backup-import-*")
	if err != nil {
		return fmt.Errorf("unable to create extraction directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	common.LogStderr(fmt.Sprintf("Extracting backup %s", backupFile))
	if err := TarGzExtract(backupFile, workDir); err != nil {
		return err
	}

	manifest, err := ReadManifest(filepath.Join(workDir, ManifestFilename))
	if err != nil {
		return fmt.Errorf("unable to read backup manifest: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return err
	}
	if manifest.DokkuVersion != "" && manifest.DokkuVersion != dokkuVersion() {
		common.LogWarn(fmt.Sprintf("backup was created with dokku %s; restoring onto %s", manifest.DokkuVersion, dokkuVersion()))
	}

	explicitApp := len(appNames) > 0
	explicitService := len(serviceSpecs) > 0
	var apps []string
	var services []serviceRef
	if !explicitApp && !explicitService {
		apps = selectApps(manifest, nil)
		services = selectServices(manifest, nil)
	} else {
		if explicitApp {
			apps = selectApps(manifest, appNames)
		}
		if explicitService {
			services = selectServices(manifest, serviceSpecs)
		}
	}

	conflicts := preflightConflicts(apps, services, force)
	if len(conflicts) > 0 {
		return fmt.Errorf("the following already exist and would be overwritten; re-run with --force to replace them: %s", strings.Join(conflicts, ", "))
	}

	var failures []string

	// Reinstall third-party plugins first, before any other restore step, so
	// that datastore and other plugins exist before their state is restored.
	if installPlugins {
		os.Setenv("DOKKU_BACKUP_INSTALL_PLUGINS", "1")
		common.LogStderr("Reinstalling plugins recorded in the backup")
		if err := dispatchTrigger("backup-plugins-install", globalScopeDir(workDir)); err != nil {
			common.LogWarn(fmt.Sprintf("plugin reinstall failed: %v", err))
			failures = append(failures, "plugins")
		}
	}

	if err := dispatchTrigger("pre-backup-import", workDir, backupFile); err != nil {
		return err
	}

	if err := dispatchTrigger("backup-global-import", globalScopeDir(workDir)); err != nil {
		common.LogWarn(fmt.Sprintf("global import failed: %v", err))
		failures = append(failures, "global")
	}

	for _, ref := range services {
		if err := importService(workDir, ref, force); err != nil {
			common.LogWarn(err.Error())
			failures = append(failures, ref.String())
		}
	}

	for _, appName := range apps {
		if err := importApp(workDir, appName, manifest.Apps[appName], force); err != nil {
			common.LogWarn(err.Error())
			failures = append(failures, appName)
		}
	}

	if err := dispatchTrigger("post-backup-import", workDir, backupFile); err != nil {
		return err
	}

	if len(failures) > 0 {
		return fmt.Errorf("backup import completed with errors for: %s", strings.Join(failures, ", "))
	}
	common.LogStderr("Backup import complete")
	return nil
}

// selectApps returns the apps from the manifest, optionally filtered to the
// requested subset.
func selectApps(manifest *Manifest, appNames []string) []string {
	var apps []string
	if len(appNames) == 0 {
		for name := range manifest.Apps {
			apps = append(apps, name)
		}
	} else {
		for _, name := range appNames {
			if _, ok := manifest.Apps[name]; ok {
				apps = append(apps, name)
			}
		}
	}
	sort.Strings(apps)
	return apps
}

// selectServices returns the services from the manifest, optionally filtered to
// the requested subset (by "type:name" or a bare "type").
func selectServices(manifest *Manifest, serviceSpecs []string) []serviceRef {
	var refs []serviceRef
	for key := range manifest.Services {
		parts := strings.SplitN(key, "/", 2)
		if len(parts) != 2 {
			continue
		}
		ref := serviceRef{Type: parts[0], Name: parts[1]}
		if len(serviceSpecs) == 0 || serviceMatchesSpecs(ref, serviceSpecs) {
			refs = append(refs, ref)
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].key() < refs[j].key() })
	return refs
}

// serviceMatchesSpecs reports whether a service matches any of the requested
// --service specs.
func serviceMatchesSpecs(ref serviceRef, serviceSpecs []string) bool {
	for _, spec := range serviceSpecs {
		serviceType, serviceName := parseServiceSpec(spec)
		if serviceType != ref.Type {
			continue
		}
		if serviceName == "" || serviceName == ref.Name {
			return true
		}
	}
	return false
}

// preflightConflicts returns the names of existing, accessible targets that
// would be overwritten without --force. Inaccessible existing targets are not
// listed (they are skipped generically during restore).
func preflightConflicts(apps []string, services []serviceRef, force bool) []string {
	if force {
		return nil
	}
	var conflicts []string
	for _, appName := range apps {
		if appExists(appName) && callerHasAppAccess(appName) {
			conflicts = append(conflicts, appName)
		}
	}
	for _, ref := range services {
		if serviceExists(ref) && callerHasServiceAccess(ref) {
			conflicts = append(conflicts, ref.String())
		}
	}
	sort.Strings(conflicts)
	return conflicts
}

// importApp restores a single app: it enforces access, (re)creates the app,
// dispatches the app-import triggers so each plugin reapplies its slice, and
// redeploys once when the app was deployed with code.
func importApp(workDir string, appName string, meta ManifestApp, force bool) error {
	scope := appScopeDir(workDir, appName)
	exists := appExists(appName)

	if exists && !callerHasAppAccess(appName) {
		return fmt.Errorf("Unable to restore %s", appName)
	}
	if exists {
		if !force {
			return fmt.Errorf("Unable to restore %s", appName)
		}
		if err := dispatchTrigger("app-destroy", appName); err != nil {
			return fmt.Errorf("unable to replace existing app %s: %w", appName, err)
		}
	}

	if err := dispatchTrigger("app-create", appName); err != nil {
		return fmt.Errorf("unable to create app %s: %w", appName, err)
	}

	if err := dispatchTrigger("pre-backup-app-import", appName, scope); err != nil {
		return err
	}
	if err := dispatchTrigger("backup-app-import", appName, scope); err != nil {
		return err
	}
	if err := dispatchTrigger("post-backup-app-import", appName, scope); err != nil {
		return err
	}

	if meta.Deployed && meta.HasCode {
		// Give plugins a chance to act right before the restore redeploy (for
		// example refreshing a TLS certificate). Best-effort: a failure here must
		// not block the redeploy.
		if err := dispatchTrigger("pre-backup-app-deploy", appName, scope); err != nil {
			common.LogWarn(fmt.Sprintf("pre-backup-app-deploy failed for %s: %v", appName, err))
		}

		common.LogStderr(fmt.Sprintf("Redeploying %s", appName))
		if _, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger:      "receive-app",
			Args:         []string{appName},
			StreamStderr: true,
		}); err != nil {
			return fmt.Errorf("unable to redeploy %s: %w", appName, err)
		}
	}

	return nil
}

// importService restores a single service: it enforces access then dispatches
// the service-import triggers so the datastore plugin recreates it.
func importService(workDir string, ref serviceRef, force bool) error {
	scope := serviceScopeDir(workDir, ref.Type, ref.Name)
	if !common.DirectoryExists(scope) {
		return nil
	}

	exists := serviceExists(ref)
	if exists && !callerHasServiceAccess(ref) {
		return fmt.Errorf("Unable to restore %s", ref)
	}
	if exists && !force {
		return fmt.Errorf("Unable to restore %s", ref)
	}

	if err := dispatchTrigger("pre-backup-service-import", ref.Type, ref.Name, scope); err != nil {
		return err
	}
	if err := dispatchTrigger("backup-service-import", ref.Type, ref.Name, scope); err != nil {
		return err
	}
	return dispatchTrigger("post-backup-service-import", ref.Type, ref.Name, scope)
}
