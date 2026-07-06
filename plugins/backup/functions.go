package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// serviceRef identifies a single datastore service by type and name.
type serviceRef struct {
	Type string
	Name string
}

// String renders a service reference as type:name.
func (s serviceRef) String() string {
	return fmt.Sprintf("%s:%s", s.Type, s.Name)
}

// key returns the manifest key for a service ("type/name").
func (s serviceRef) key() string {
	return fmt.Sprintf("%s/%s", s.Type, s.Name)
}

// parseServiceSpec parses a --service value. "type:name" selects one service;
// a bare "type" selects every service of that type.
func parseServiceSpec(spec string) (serviceType string, serviceName string) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// sshUser returns the authenticated SSH user, mirroring filterApps.
func sshUser() string {
	if user := os.Getenv("SSH_USER"); user != "" {
		return user
	}
	return os.Getenv("USER")
}

// sshName returns the authenticated SSH key name, mirroring filterApps.
func sshName() string {
	if name := os.Getenv("SSH_NAME"); name != "" {
		return name
	}
	if name := os.Getenv("NAME"); name != "" {
		return name
	}
	return "default"
}

// uniqueNonEmptyLines splits text on newlines and returns the unique,
// whitespace-trimmed, non-empty lines in stable order.
func uniqueNonEmptyLines(text string) []string {
	seen := map[string]bool{}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		lines = append(lines, line)
	}
	return lines
}

// discoverServiceTypes broadcasts the datastore-list trigger and returns the
// unique set of installed datastore types. Each datastore plugin echoes its own
// type in response.
func discoverServiceTypes() ([]string, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "datastore-list",
		Env:     streamTriggerEnv,
	})
	if err != nil {
		return nil, err
	}
	return uniqueNonEmptyLines(result.StdoutContents()), nil
}

// listServicesOfType calls the type-scoped service-list trigger and returns the
// services of the requested type as type:name references.
func listServicesOfType(serviceType string) ([]serviceRef, error) {
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "service-list",
		Args:    []string{serviceType},
		Env:     streamTriggerEnv,
	})
	if err != nil {
		return nil, err
	}

	var refs []serviceRef
	for _, line := range uniqueNonEmptyLines(result.StdoutContents()) {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		refs = append(refs, serviceRef{Type: parts[0], Name: parts[1]})
	}
	return refs, nil
}

// discoverAllServices enumerates every datastore type via datastore-list, lists
// each type's services via service-list, and returns the access-filtered,
// deduplicated set.
func discoverAllServices() ([]serviceRef, error) {
	types, err := discoverServiceTypes()
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var refs []serviceRef
	for _, serviceType := range types {
		typeRefs, err := listServicesOfType(serviceType)
		if err != nil {
			return nil, err
		}
		for _, ref := range typeRefs {
			if seen[ref.key()] || !callerHasServiceAccess(ref) {
				continue
			}
			seen[ref.key()] = true
			refs = append(refs, ref)
		}
	}

	sort.Slice(refs, func(i, j int) bool { return refs[i].key() < refs[j].key() })
	return refs, nil
}

// callerHasAppAccess reports whether the SSH caller may access an app. It reuses
// the user-auth-app trigger (the same one that filters common.DokkuApps); when
// no ACL plugin provides it, all access is granted.
func callerHasAppAccess(appName string) bool {
	if !common.PlugnTriggerExists("user-auth-app") {
		return true
	}
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "user-auth-app",
		Args:    []string{sshUser(), sshName(), appName},
		Env:     streamTriggerEnv,
	})
	if err != nil {
		return false
	}
	for _, line := range uniqueNonEmptyLines(result.StdoutContents()) {
		if line == appName {
			return true
		}
	}
	return false
}

// callerHasServiceAccess reports whether the SSH caller may access a service.
// There is no service list-filter trigger, so it reuses the generic user-auth
// command-authorization trigger with the service's canonical export command.
// When no ACL plugin provides user-auth, all access is granted.
func callerHasServiceAccess(ref serviceRef) bool {
	if !common.PlugnTriggerExists("user-auth") {
		return true
	}
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "user-auth",
		Args:    []string{sshUser(), sshName(), fmt.Sprintf("%s:export", ref.Type), ref.Name},
		Env:     streamTriggerEnv,
	})
	if err != nil {
		return false
	}
	return result.ExitCode == 0
}

// appExists reports whether an app currently exists on the host. It checks the
// app root directly rather than via the access-filtered app-exists trigger, so
// an app the caller cannot access still registers as existing and is never
// silently treated as new (and clobbered) during import.
func appExists(appName string) bool {
	return common.DirectoryExists(common.AppRoot(appName))
}

// serviceExists reports whether a service currently exists on the host, by
// re-listing its type and matching the name.
func serviceExists(ref serviceRef) bool {
	refs, err := listServicesOfType(ref.Type)
	if err != nil {
		return false
	}
	for _, existing := range refs {
		if existing.Name == ref.Name {
			return true
		}
	}
	return false
}

// ensureBackupDirWritable verifies the backup directory exists and is writable
// by creating and removing a probe file. It fails fast before any work is done.
func ensureBackupDirWritable(backupDir string) error {
	info, err := os.Stat(backupDir)
	if err != nil {
		return fmt.Errorf("backup-dir does not exist: %s", backupDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("backup-dir is not a directory: %s", backupDir)
	}
	probe, err := os.CreateTemp(backupDir, ".dokku-backup-write-test-*")
	if err != nil {
		return fmt.Errorf("backup-dir is not writable: %s: %w", backupDir, err)
	}
	probe.Close()
	return os.Remove(probe.Name())
}

// setOwnership best-effort sets dokku:dokku ownership recursively on a path.
func setOwnership(path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		mode := os.FileMode(0600)
		if info.IsDir() {
			mode = 0755
		}
		return common.SetPermissions(common.SetPermissionInput{
			Filename: p,
			Mode:     mode,
		})
	})
}
