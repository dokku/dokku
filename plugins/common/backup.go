package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	yaml "go.yaml.in/yaml/v3"
)

// BackupRecipe is a docket-compatible recipe: an ordered list of plays. A
// backup serializes each plugin's declarative state as one of these so the
// resulting files are valid docket recipes that can also be applied externally
// with `docket apply` (docket >= 0.6.0).
type BackupRecipe []BackupPlay

// BackupPlay is a single docket play: a named list of tasks.
type BackupPlay struct {
	// Name is a human label for the play, shown in docket output.
	Name string `yaml:"name,omitempty"`

	// Tasks is the ordered list of tasks in the play. Each task is a single
	// dokku_* key mapping to its argument body.
	Tasks []BackupTask `yaml:"tasks"`
}

// BackupTask is a single docket task: exactly one dokku_* key mapping to its
// argument body. Bespoke plugins place a typed struct as the value to control
// field ordering; the generic property helper places a BackupPropertyTask.
type BackupTask map[string]interface{}

// BackupPropertyTask is the body of a generic dokku_<plugin>_property task,
// matching the docket property-task schema shared by every plugin.
type BackupPropertyTask struct {
	// App is the name of the app. Empty when Global is true.
	App string `yaml:"app,omitempty"`

	// Global indicates the property applies to the global scope.
	Global bool `yaml:"global,omitempty"`

	// Property is the name of the property to set.
	Property string `yaml:"property"`

	// Value is the value to set for the property.
	Value string `yaml:"value,omitempty"`

	// State is the desired state of the property (present|absent).
	State string `yaml:"state,omitempty"`
}

// globalScopeName is the app name used to denote the global property scope.
const globalScopeName = "--global"

// BackupConfigDir returns the config sub-tree of a backup scope directory,
// where plugins write their docket recipe slices.
func BackupConfigDir(scopeDir string) string {
	return filepath.Join(scopeDir, "config")
}

// BackupDataDir returns the data sub-tree of a backup scope directory, where
// plugins write free-form bulk data (tarballs, git bundles, cert files).
func BackupDataDir(scopeDir string) string {
	return filepath.Join(scopeDir, "data")
}

// BackupRecipePath returns the path of a plugin's recipe slice within a scope
// directory. The slice lives under the scope's config sub-tree as
// <scopeDir>/config/<pluginName>.yml.
func BackupRecipePath(pluginName string, scopeDir string) string {
	return filepath.Join(BackupConfigDir(scopeDir), pluginName+".yml")
}

// BackupWriteRecipe marshals a recipe to the given path, creating the parent
// directory and setting dokku:dokku ownership with 0600 permissions.
func BackupWriteRecipe(path string, recipe BackupRecipe) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("unable to create recipe directory: %w", err)
	}

	data, err := yaml.Marshal(recipe)
	if err != nil {
		return fmt.Errorf("unable to marshal recipe: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("unable to write recipe: %w", err)
	}

	return SetPermissions(SetPermissionInput{
		Filename: path,
		Mode:     0600,
	})
}

// BackupReadRecipeFile reads a recipe slice from disk and unmarshals it into
// the caller-provided typed destination. Returns os.ErrNotExist (wrapped) when
// the file is absent so callers can treat a missing slice as a no-op.
func BackupReadRecipeFile(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("unable to parse recipe %s: %w", path, err)
	}

	return nil
}

// BackupPropertyExport serializes every property of a plugin (for one app, or
// for the global scope when appName is "--global") into a docket recipe slice
// at <scopeDir>/config/<pluginName>.yml. taskType overrides the emitted task
// key; when empty it defaults to dokku_<pluginName>_property. A plugin with no
// properties writes nothing.
func BackupPropertyExport(pluginName string, appName string, scopeDir string, taskType string) error {
	return BackupPropertyExportFiltered(pluginName, appName, scopeDir, taskType, nil)
}

// BackupPropertyExportFiltered is BackupPropertyExport with a set of property
// keys excluded from the slice. Use it to drop transient/derived properties
// that must not be restored (for example ps "scale.old" or ports "map-detected").
func BackupPropertyExportFiltered(pluginName string, appName string, scopeDir string, taskType string, exclude []string) error {
	if taskType == "" {
		taskType = fmt.Sprintf("dokku_%s_property", pluginName)
	}

	properties, err := PropertyGetAll(pluginName, appName)
	if err != nil {
		return err
	}

	excluded := map[string]bool{}
	for _, key := range exclude {
		excluded[key] = true
	}

	keys := make([]string, 0, len(properties))
	for key := range properties {
		if excluded[key] {
			continue
		}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)

	global := appName == globalScopeName
	tasks := make([]BackupTask, 0, len(keys))
	for _, key := range keys {
		body := BackupPropertyTask{
			Property: key,
			Value:    properties[key],
		}
		if global {
			body.Global = true
		} else {
			body.App = appName
		}
		tasks = append(tasks, BackupTask{taskType: body})
	}

	recipe := BackupRecipe{{
		Name:  fmt.Sprintf("%s properties", pluginName),
		Tasks: tasks,
	}}

	return BackupWriteRecipe(BackupRecipePath(pluginName, scopeDir), recipe)
}

// backupPropertyRecipe is the read shape for a generic property recipe slice:
// each task is a single-key map from the task type to its property body.
type backupPropertyRecipe []struct {
	Tasks []map[string]BackupPropertyTask `yaml:"tasks"`
}

// BackupPropertyImport reads a plugin's recipe slice from <scopeDir>/config/
// <pluginName>.yml and reapplies each property natively via the property store.
// A property whose state is "absent" is deleted; everything else is written.
// A missing slice is a no-op. The reapply is idempotent: importing the same
// slice twice converges to the same state.
func BackupPropertyImport(pluginName string, appName string, scopeDir string) error {
	path := BackupRecipePath(pluginName, scopeDir)
	if !FileExists(path) {
		return nil
	}

	var recipe backupPropertyRecipe
	if err := BackupReadRecipeFile(path, &recipe); err != nil {
		return err
	}

	for _, play := range recipe {
		for _, task := range play.Tasks {
			for _, body := range task {
				if body.Property == "" {
					continue
				}
				if body.State == "absent" {
					if err := PropertyDelete(pluginName, appName, body.Property); err != nil {
						return err
					}
					continue
				}
				if err := PropertyWrite(pluginName, appName, body.Property, body.Value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
