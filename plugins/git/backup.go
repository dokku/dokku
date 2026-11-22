package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// backupBundlePath returns where an app's git bundle lives within a backup scope dir.
func backupBundlePath(scopeDir string) string {
	return filepath.Join(common.BackupDataDir(scopeDir), "repo.bundle")
}

// TriggerBackupAppExport exports an app's git properties and bundles its
// repository into the backup scope dir.
func TriggerBackupAppExport(appName string, scopeDir string) error {
	if err := common.BackupPropertyExport("git", appName, scopeDir, ""); err != nil {
		return err
	}

	appRoot := common.AppRoot(appName)
	if !common.DirectoryExists(filepath.Join(appRoot, "refs")) {
		return nil
	}

	refs, err := runGit(appRoot, "for-each-ref")
	if err != nil || refs == "" {
		return nil
	}

	bundle := backupBundlePath(scopeDir)
	if err := os.MkdirAll(filepath.Dir(bundle), 0755); err != nil {
		return err
	}

	_, err = runGit(appRoot, "bundle", "create", bundle, "--all")
	return err
}

// TriggerBackupAppImport restores an app's git properties and repository from
// the backup scope dir. Tags are fetched best-effort, since a bundle without
// tags fails that refspec.
func TriggerBackupAppImport(appName string, scopeDir string) error {
	if err := common.BackupPropertyImport("git", appName, scopeDir); err != nil {
		return err
	}

	bundle := backupBundlePath(scopeDir)
	if !common.FileExists(bundle) {
		return nil
	}

	appRoot := common.AppRoot(appName)
	if err := ensureReceiveRepository(appName, appRoot); err != nil {
		return err
	}

	if _, err := runGit(appRoot, "fetch", bundle, "+refs/heads/*:refs/heads/*"); err != nil {
		return err
	}
	_, _ = runGit(appRoot, "fetch", bundle, "+refs/tags/*:refs/tags/*")

	return common.BackupChownTree(appRoot)
}

// ensureReceiveRepository initializes the app's bare repository when missing
// and writes the pre-receive hook that hands pushes to dokku, matching
// fn-git-create-hook.
func ensureReceiveRepository(appName string, appRoot string) error {
	if !common.DirectoryExists(filepath.Join(appRoot, "refs")) {
		if _, err := runGit("", "init", "-q", "--bare", appRoot); err != nil {
			return err
		}
	}

	hook := fmt.Sprintf(`#!/usr/bin/env bash
set -e
set -o pipefail

cat | DOKKU_ROOT="%s" dokku git-hook %s
`, common.MustGetEnv("DOKKU_ROOT"), appName)

	hookPath := filepath.Join(appRoot, "hooks", "pre-receive")
	if err := os.WriteFile(hookPath, []byte(hook), 0755); err != nil {
		return err
	}
	return os.Chmod(hookPath, 0755)
}

// runGit runs a git command in the given directory and returns its trimmed
// stdout, failing on a non-zero exit.
func runGit(dir string, args ...string) (string, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:          "git",
		Args:             args,
		WorkingDirectory: dir,
	})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), result.StderrContents())
	}

	return strings.TrimSpace(result.StdoutContents()), nil
}
