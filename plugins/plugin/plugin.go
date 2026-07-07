package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokku/dokku/plugins/common"

	"golang.org/x/sync/errgroup"
)

// PluginInfo describes an installed dokku plugin
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Enabled     bool   `json:"enabled"`
	Core        bool   `json:"core"`
	Description string `json:"description"`
	SourceURL   string `json:"source_url"`
	Committish  string `json:"committish"`
	Branch      string `json:"branch"`
}

// CommandList lists all installed plugins in the specified format
func CommandList(format string) error {
	if format != "json" {
		return fmt.Errorf("Invalid output format specified, supported formats: json")
	}

	plugins, err := listPlugins()
	if err != nil {
		return err
	}

	out, err := json.Marshal(plugins)
	if err != nil {
		return err
	}

	common.Log(string(out))
	return nil
}

// listPlugins returns metadata for every installed plugin, enriched with git
// source information for git-based third-party plugins
func listPlugins() ([]PluginInfo, error) {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "plugn",
		Args:    []string{"list"},
	})
	if err != nil {
		return nil, err
	}

	coreEnabledPath := common.MustGetEnv("PLUGIN_CORE_ENABLED_PATH")
	availablePath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")

	plugins := parsePlugnList(result.StdoutContents())

	errs := new(errgroup.Group)
	for i := range plugins {
		i := i
		errs.Go(func() error {
			plugins[i].Core = common.IsSymlink(filepath.Join(coreEnabledPath, plugins[i].Name))
			enrichGitSource(&plugins[i], availablePath)
			return nil
		})
	}
	if err := errs.Wait(); err != nil {
		return nil, err
	}

	return plugins, nil
}

// parsePlugnList parses the output of `plugn list` into plugin metadata. Each
// data line is formatted as `<name> <version> <enabled|disabled> <description>`,
// preceded by a `plugn: <treeish>` header line that is skipped.
func parsePlugnList(output string) []PluginInfo {
	plugins := []PluginInfo{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] == "plugn:" {
			continue
		}

		description := ""
		if len(fields) > 3 {
			description = strings.Join(fields[3:], " ")
		}

		plugins = append(plugins, PluginInfo{
			Name:        fields[0],
			Version:     fields[1],
			Enabled:     fields[2] == "enabled",
			Description: description,
		})
	}

	return plugins
}

// enrichGitSource populates the git source fields for a git-based plugin,
// leaving them empty for core, tarball, or file-based installs
func enrichGitSource(info *PluginInfo, availablePath string) {
	pluginDir := filepath.Join(availablePath, info.Name)
	if _, err := os.Stat(filepath.Join(pluginDir, ".git")); err != nil {
		return
	}

	info.SourceURL = gitOutput(pluginDir, "remote", "get-url", "origin")
	info.Committish = gitOutput(pluginDir, "rev-parse", "HEAD")
	if branch := gitOutput(pluginDir, "rev-parse", "--abbrev-ref", "HEAD"); branch != "HEAD" {
		info.Branch = branch
	}
}

// gitOutput runs a git command in the given directory and returns its trimmed
// stdout, or an empty string on error. The directory is marked as safe so the
// lookup works regardless of which user invokes plugin:list.
func gitOutput(dir string, args ...string) string {
	gitArgs := append([]string{"-c", "safe.directory=" + dir}, args...)
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:          "git",
		Args:             gitArgs,
		WorkingDirectory: dir,
	})
	if err != nil {
		return ""
	}

	return result.StdoutContents()
}
