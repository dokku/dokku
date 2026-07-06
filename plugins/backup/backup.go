// Package backup implements the dokku core backup plugin. It provides the
// backup:export and backup:import commands, which serialize an app (and/or
// service, and/or the global scope) into a single tar.gz and restore it again.
//
// Core owns only orchestration: scope resolution, the manifest, the staging
// directory lifecycle, tar/gzip, ordering, and stdout/stderr discipline. Every
// plugin serializes and reapplies its own slice of state through the
// backup-*-export / backup-*-import plugin triggers, so core never reaches into
// another plugin's files. Declarative config is written as docket recipe
// slices (docket >= 0.6.0); non-declarative state (git bundle, certs, storage)
// is written as raw files under each scope's data/ sub-tree.
package backup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	// ManifestFilename is the name of the manifest within a backup.
	ManifestFilename = "manifest.json"

	// ManifestSchemaVersion is the current manifest schema version. Import
	// refuses backups whose schema_version is newer than this.
	ManifestSchemaVersion = 1

	// BackupFormatVersion identifies the tarball format.
	BackupFormatVersion = "dokku-backup/1"

	// GlobalDir is the scope directory for global state within a backup.
	GlobalDir = "global"

	// AppsDir is the parent directory for per-app scopes within a backup.
	AppsDir = "apps"

	// ServicesDir is the parent directory for per-service scopes within a backup.
	ServicesDir = "services"

	// GlobalScopeName is the app name used to denote the global property scope.
	GlobalScopeName = "--global"
)

// appScopeDir returns the scope directory for an app within a backup root.
func appScopeDir(root string, appName string) string {
	return filepath.Join(root, AppsDir, appName)
}

// serviceScopeDir returns the scope directory for a service within a backup root.
func serviceScopeDir(root string, serviceType string, serviceName string) string {
	return filepath.Join(root, ServicesDir, serviceType, serviceName)
}

// globalScopeDir returns the global scope directory within a backup root.
func globalScopeDir(root string) string {
	return filepath.Join(root, GlobalDir)
}

// dokkuVersion returns the installed dokku version string, or "unknown".
func dokkuVersion() string {
	libRoot := os.Getenv("DOKKU_LIB_ROOT")
	if libRoot == "" {
		libRoot = "/var/lib/dokku"
	}
	for _, name := range []string{"STABLE_VERSION", "VERSION"} {
		data, err := os.ReadFile(filepath.Join(libRoot, name))
		if err == nil {
			if version := strings.TrimSpace(string(data)); version != "" {
				return version
			}
		}
	}
	return "unknown"
}

// dokkuLibRoot returns the dokku data root, defaulting to /var/lib/dokku.
func dokkuLibRoot() string {
	if libRoot := os.Getenv("DOKKU_LIB_ROOT"); libRoot != "" {
		return libRoot
	}
	return "/var/lib/dokku"
}

// hostname returns the host name for the manifest, or "" on error.
func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

// streamTriggerEnv is the env applied to dispatched sub-triggers so their
// human output is suppressed and never leaks onto our stdout.
var streamTriggerEnv = map[string]string{"DOKKU_QUIET_OUTPUT": "1"}

// dispatchTrigger runs a plugn trigger for every plugin that implements it,
// keeping the trigger's stdout off our own stdout (only the final tarball path
// is written to stdout). Any captured stdout is re-emitted on stderr.
func dispatchTrigger(trigger string, args ...string) error {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:      trigger,
		Args:         args,
		Env:          streamTriggerEnv,
		StreamStdout: false,
		StreamStderr: true,
	})
	if err != nil {
		return err
	}
	if out := strings.TrimSpace(result.StdoutContents()); out != "" {
		common.LogStderr(out)
	}
	return nil
}

// dispatchTriggerWithEnv is dispatchTrigger with additional environment
// variables merged into the trigger invocation.
func dispatchTriggerWithEnv(env map[string]string, trigger string, args ...string) error {
	merged := map[string]string{"DOKKU_QUIET_OUTPUT": "1"}
	for key, value := range env {
		merged[key] = value
	}
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:      trigger,
		Args:         args,
		Env:          merged,
		StreamStdout: false,
		StreamStderr: true,
	})
	if err != nil {
		return err
	}
	if out := strings.TrimSpace(result.StdoutContents()); out != "" {
		common.LogStderr(out)
	}
	return nil
}
