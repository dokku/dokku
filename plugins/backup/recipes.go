package backup

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/dokku/dokku/plugins/common"
	yaml "go.yaml.in/yaml/v3"
)

// assembleRecipes writes the combined docket recipes that make a backup
// applicable with `docket apply` (docket >= 0.6.0). It produces a per-scope
// tasks.yml for the global scope, each app, and each service, plus a whole-
// backup tasks.yml at the root combining them in apply order: global first,
// then apps, then services. These files are not read by import (each plugin
// reapplies its own slice); they exist for external, out-of-band restore.
func assembleRecipes(root string, apps []string, services []serviceRef) error {
	var combined []yaml.Node

	globalDir := globalScopeDir(root)
	globalPlays, err := combineScope(globalDir)
	if err != nil {
		return err
	}
	combined = append(combined, globalPlays...)

	for _, appName := range apps {
		plays, err := combineScope(appScopeDir(root, appName))
		if err != nil {
			return err
		}
		combined = append(combined, plays...)
	}

	for _, ref := range services {
		plays, err := combineScope(serviceScopeDir(root, ref.Type, ref.Name))
		if err != nil {
			return err
		}
		combined = append(combined, plays...)
	}

	return writePlays(filepath.Join(root, "tasks.yml"), combined)
}

// combineScope reads every recipe slice in a scope's config directory, in
// stable filename order, and writes their concatenation to the scope's
// tasks.yml. It returns the combined plays so callers can roll them up.
func combineScope(scopeDir string) ([]yaml.Node, error) {
	configDir := common.BackupConfigDir(scopeDir)
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, nil
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) == ".yml" && name != "tasks.yml" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var plays []yaml.Node
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(configDir, name))
		if err != nil {
			return nil, err
		}
		var slice []yaml.Node
		if err := yaml.Unmarshal(data, &slice); err != nil {
			continue
		}
		plays = append(plays, slice...)
	}

	if len(plays) == 0 {
		return nil, nil
	}

	if err := writePlays(filepath.Join(scopeDir, "tasks.yml"), plays); err != nil {
		return nil, err
	}
	return plays, nil
}

// writePlays marshals a list of plays to a recipe file with 0600 permissions.
func writePlays(path string, plays []yaml.Node) error {
	if len(plays) == 0 {
		return nil
	}
	data, err := yaml.Marshal(plays)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
