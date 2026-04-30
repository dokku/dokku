package builds

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall sets up the builds plugin's property storage.
func TriggerInstall() error {
	if err := common.PropertySetup("builds"); err != nil {
		return fmt.Errorf("Unable to install the builds plugin: %s", err.Error())
	}
	return nil
}

// TriggerPostDelete removes all builds data and properties for the given app.
func TriggerPostDelete(appName string) error {
	if err := os.RemoveAll(AppDataDir(appName)); err != nil {
		common.LogWarn(fmt.Sprintf("Could not remove builds data for %s: %s", appName, err))
	}
	if err := common.PropertyDestroy("builds", appName); err != nil {
		return err
	}
	return nil
}

// TriggerPostAppRenameSetup renames the per-app data directory and clones
// builds properties to the new app name.
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	oldDir := AppDataDir(oldAppName)
	newDir := AppDataDir(newAppName)
	if _, err := os.Stat(oldDir); err == nil {
		if err := os.MkdirAll(common.GetAppDataDirectory("builds", ""), 0755); err == nil {
			if err := os.Rename(oldDir, newDir); err != nil {
				common.LogWarn(fmt.Sprintf("Could not rename builds data dir for %s: %s", oldAppName, err))
			}
		}
	}

	if err := common.PropertyClone("builds", oldAppName, newAppName); err != nil {
		return err
	}
	if err := common.PropertyDestroy("builds", oldAppName); err != nil {
		return err
	}
	return nil
}

// TriggerBuildsGenerateID writes a fresh build-id to stdout. The bash callers
// of this trigger consume the value via $(...) capture, so this function MUST
// emit only the build-id and nothing else - no log output, no warnings.
func TriggerBuildsGenerateID() error {
	fmt.Println(GenerateBuildID())
	return nil
}

// TriggerBuildsRecordStart persists the initial build record on lock-acquire.
//
// Args: <app> <build-id> <pid> <source>
//
// kind is derived from source via BuildSource.DefaultKind(); callers do not
// pass it.
func TriggerBuildsRecordStart(appName, buildID, pidStr, sourceStr string) error {
	if appName == "" {
		return errors.New("builds-record-start: missing app name")
	}
	if buildID == "" {
		return errors.New("builds-record-start: missing build id")
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("builds-record-start: invalid pid %q: %w", pidStr, err)
	}

	source := BuildSource(sourceStr)
	if !source.Valid() {
		common.LogWarn(fmt.Sprintf("builds-record-start: unknown source %q for app %s, recording as %q", sourceStr, appName, BuildSourceUnknown))
		source = BuildSourceUnknown
	}

	b := Build{
		ID:        buildID,
		App:       appName,
		Kind:      source.DefaultKind(),
		PID:       pid,
		StartedAt: time.Now().UTC(),
		Status:    BuildStatusRunning,
		Source:    source,
	}
	return WriteBuild(b)
}

// TriggerBuildsRecordFinalize writes the terminal status onto an existing
// build record. It is idempotent: records that are already terminal (or have
// been overwritten by an earlier finalize) are left untouched.
//
// Args: <app> <build-id> <exit-code>
func TriggerBuildsRecordFinalize(appName, buildID, exitStr string) error {
	if appName == "" {
		return errors.New("builds-record-finalize: missing app name")
	}
	if buildID == "" {
		return errors.New("builds-record-finalize: missing build id")
	}

	exitCode, err := strconv.Atoi(exitStr)
	if err != nil {
		return fmt.Errorf("builds-record-finalize: invalid exit code %q: %w", exitStr, err)
	}

	b, err := ReadBuild(appName, buildID)
	if err != nil {
		if os.IsNotExist(err) {
			common.LogWarn(fmt.Sprintf("builds-record-finalize: no record for %s/%s, skipping", appName, buildID))
			return nil
		}
		return err
	}

	if b.Status.IsTerminal() {
		// Idempotent path - typically hit when builds:cancel finalized the
		// record before the dying process reached release_app_deploy_lock.
		return PruneAppBuilds(appName)
	}

	now := time.Now().UTC()
	b.FinishedAt = &now
	b.ExitCode = &exitCode
	if exitCode == 0 {
		b.Status = BuildStatusSucceeded
	} else {
		b.Status = BuildStatusFailed
	}
	if err := WriteBuild(b); err != nil {
		return err
	}

	return PruneAppBuilds(appName)
}
