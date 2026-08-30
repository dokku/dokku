package scheduler_k3s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"helm.sh/helm/v3/pkg/chartutil"
)

// NodeSysctlsValues contains the values for a dokku-managed node sysctls helm chart
type NodeSysctlsValues struct {
	Global NodeSysctlsGlobalValues `yaml:"global"`
}

// NodeSysctlsGlobalValues contains the global values for the node sysctls chart
type NodeSysctlsGlobalValues struct {
	Image       string   `yaml:"image"`
	PauseImage  string   `yaml:"pause_image"`
	ProfileName string   `yaml:"profile_name,omitempty"`
	ReleaseName string   `yaml:"release_name"`
	Sysctls     []Sysctl `yaml:"sysctls"`
}

// nodeSysctlScope is a resolved set of sysctls destined for a single DaemonSet
type nodeSysctlScope struct {
	// ProfileName is the node profile this scope targets, empty for the global scope
	ProfileName string
	// ReleaseName is the helm release backing this scope
	ReleaseName string
	// Sysctls are the fully resolved sysctls for this scope
	Sysctls []Sysctl
}

// getNodeSysctlsProperty returns the property name backing a node sysctls scope
func getNodeSysctlsProperty(profileName string) string {
	if profileName == "" {
		return "node-sysctls.global"
	}

	return fmt.Sprintf("node-sysctls.profile.%s", profileName)
}

// nodeSysctlsProfileReleasePrefix prefixes the helm release name of every profile-scoped
// node sysctls chart. A profile name is appended to it verbatim, so this is what bounds
// how long a profile name may be.
const nodeSysctlsProfileReleasePrefix = "dokku-node-sysctls-profile-"

// getNodeSysctlsReleaseName returns the helm release name for a node sysctls scope
func getNodeSysctlsReleaseName(profileName string) string {
	if profileName == "" {
		return "dokku-node-sysctls-global"
	}

	return nodeSysctlsProfileReleasePrefix + profileName
}

// getNodeSysctls returns the sysctls stored against a single scope
func getNodeSysctls(profileName string) (map[string]string, error) {
	sysctls, err := common.PropertyMapGet("scheduler-k3s", "--global", getNodeSysctlsProperty(profileName))
	if err != nil {
		return nil, fmt.Errorf("Unable to read node sysctls: %w", err)
	}

	return sysctls, nil
}

// listNodeProfileNames returns the names of every stored node profile
func listNodeProfileNames() ([]string, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", "--global", "node-profile-")
	if err != nil {
		return nil, fmt.Errorf("Unable to get node profiles: %w", err)
	}

	names := []string{}
	for property, data := range properties {
		if !strings.HasSuffix(property, ".json") {
			continue
		}

		var profile NodeProfile
		if err := json.Unmarshal([]byte(data), &profile); err != nil {
			return nil, fmt.Errorf("Unable to unmarshal node profile: %w", err)
		}

		names = append(names, profile.Name)
	}

	sort.Strings(names)
	return names, nil
}

// sortedSysctls converts a name/value map into sysctls sorted by name so the
// rendered manifest is stable across runs
func sortedSysctls(values map[string]string) []Sysctl {
	sysctls := make([]Sysctl, 0, len(values))
	for name, value := range values {
		sysctls = append(sysctls, Sysctl{Name: name, Value: value})
	}

	sort.Slice(sysctls, func(i int, j int) bool {
		return sysctls[i].Name < sysctls[j].Name
	})

	return sysctls
}

// mergeNodeSysctls layers a profile's sysctls over the global ones, with the profile
// winning on conflict, and returns the result sorted by name. Profile scopes must
// carry the global values too: profiled nodes are excluded from the global DaemonSet,
// so anything omitted here would never reach them.
func mergeNodeSysctls(global map[string]string, profile map[string]string) []Sysctl {
	merged := map[string]string{}
	for name, value := range global {
		merged[name] = value
	}
	for name, value := range profile {
		merged[name] = value
	}

	return sortedSysctls(merged)
}

// resolveNodeSysctlScopes returns one scope per DaemonSet that should exist, with
// profile scopes carrying the global sysctls merged underneath their own. Every
// node matches exactly one scope: profiled nodes match their profile's DaemonSet,
// and unprofiled nodes match the global one, so no two DaemonSets ever write the
// same sysctl on the same node.
func resolveNodeSysctlScopes() ([]nodeSysctlScope, error) {
	globalSysctls, err := getNodeSysctls("")
	if err != nil {
		return nil, err
	}

	scopes := []nodeSysctlScope{
		{
			ReleaseName: getNodeSysctlsReleaseName(""),
			Sysctls:     sortedSysctls(globalSysctls),
		},
	}

	profileNames, err := listNodeProfileNames()
	if err != nil {
		return nil, err
	}

	for _, profileName := range profileNames {
		profileSysctls, err := getNodeSysctls(profileName)
		if err != nil {
			return nil, err
		}

		scopes = append(scopes, nodeSysctlScope{
			ProfileName: profileName,
			ReleaseName: getNodeSysctlsReleaseName(profileName),
			Sysctls:     mergeNodeSysctls(globalSysctls, profileSysctls),
		})
	}

	return scopes, nil
}

// CreateOrUpdateNodeSysctls reconciles every node sysctls DaemonSet against the
// stored properties. All scopes are reconciled on every call rather than only the
// mutated one, since profile scopes inherit the global sysctls at render time and
// would otherwise serve a stale value after a global change.
func CreateOrUpdateNodeSysctls(ctx context.Context) error {
	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping node sysctls sync")
		return nil
	}

	scopes, err := resolveNodeSysctlScopes()
	if err != nil {
		return err
	}

	for _, scope := range scopes {
		if err := chartutil.ValidateReleaseName(scope.ReleaseName); err != nil {
			common.LogWarn(fmt.Sprintf("Skipping node sysctls for %s, its name predates profile name validation and cannot back a helm release: %s", nodeSysctlsScopeLabel(scope.ProfileName), err))
			continue
		}

		if len(scope.Sysctls) == 0 {
			if err := deleteNodeSysctlsRelease(scope.ReleaseName); err != nil {
				return err
			}
			continue
		}

		if err := installNodeSysctlsChart(ctx, scope); err != nil {
			return err
		}
	}

	return nil
}

// installNodeSysctlsChart installs or upgrades the DaemonSet backing a single scope
func installNodeSysctlsChart(ctx context.Context, scope nodeSysctlScope) error {
	chartDir, err := os.MkdirTemp("", "dokku-node-sysctls-chart-")
	if err != nil {
		return fmt.Errorf("error creating chart directory: %w", err)
	}
	defer os.RemoveAll(chartDir)

	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), os.FileMode(0755)); err != nil {
		return fmt.Errorf("error creating chart templates directory: %w", err)
	}

	chart := &Chart{
		ApiVersion: "v2",
		AppVersion: "1.0.0",
		Name:       scope.ReleaseName,
		Icon:       "https://dokku.com/assets/dokku-logo.svg",
		Version:    "0.0.1",
	}

	err = writeYaml(WriteYamlInput{
		Object: chart,
		Path:   filepath.Join(chartDir, "Chart.yaml"),
	})
	if err != nil {
		return fmt.Errorf("error writing chart: %w", err)
	}

	b, err := templates.ReadFile("templates/node-sysctls-chart/templates/daemonset.yaml")
	if err != nil {
		return fmt.Errorf("error reading node-sysctls template: %w", err)
	}

	filename := filepath.Join(chartDir, "templates", "daemonset.yaml")
	if err := os.WriteFile(filename, b, os.FileMode(0644)); err != nil {
		return fmt.Errorf("error writing node-sysctls template: %w", err)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		common.CatFile(filename)
	}

	values := &NodeSysctlsValues{
		Global: NodeSysctlsGlobalValues{
			Image:       getComputedNodeSysctlsImage(),
			PauseImage:  getComputedNodeSysctlsPauseImage(),
			ProfileName: scope.ProfileName,
			ReleaseName: scope.ReleaseName,
			Sysctls:     scope.Sysctls,
		},
	}

	err = writeYaml(WriteYamlInput{
		Object: values,
		Path:   filepath.Join(chartDir, "values.yaml"),
	})
	if err != nil {
		return fmt.Errorf("error writing values: %w", err)
	}

	helmAgent, err := NewHelmAgent(NodeSysctlsNamespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}

	chartPath, err := filepath.Abs(chartDir)
	if err != nil {
		return fmt.Errorf("error getting chart path: %w", err)
	}

	common.LogVerboseQuiet(fmt.Sprintf("Applying node sysctls for %s", nodeSysctlsScopeLabel(scope.ProfileName)))
	err = helmAgent.InstallOrUpgradeChart(ctx, ChartInput{
		ChartPath:   chartPath,
		Namespace:   NodeSysctlsNamespace,
		ReleaseName: scope.ReleaseName,
		Wait:        false,
	})
	if err != nil {
		return fmt.Errorf("error installing node sysctls chart: %w", err)
	}

	return nil
}

// deleteNodeSysctlsRelease removes a node sysctls DaemonSet when its scope resolves
// to an empty set. The kernel values it wrote are not reverted; they persist on the
// affected nodes until those nodes reboot.
func deleteNodeSysctlsRelease(releaseName string) error {
	if err := chartutil.ValidateReleaseName(releaseName); err != nil {
		common.LogDebug(fmt.Sprintf("skipping node sysctls release deletion, helm could never have installed %s: %s", releaseName, err))
		return nil
	}

	helmAgent, err := NewHelmAgent(NodeSysctlsNamespace, DeployLogPrinter)
	if err != nil {
		return fmt.Errorf("error creating helm agent: %w", err)
	}

	exists, err := helmAgent.ChartExists(releaseName)
	if err != nil {
		return fmt.Errorf("error checking if node sysctls chart exists: %w", err)
	}

	if !exists {
		return nil
	}

	common.LogVerboseQuiet(fmt.Sprintf("Removing node sysctls release %s", releaseName))
	if err := helmAgent.UninstallChart(releaseName); err != nil {
		return fmt.Errorf("error uninstalling node sysctls chart: %w", err)
	}

	return nil
}

// DeleteNodeSysctls removes the stored sysctls and DaemonSet for a single node profile
func DeleteNodeSysctls(ctx context.Context, profileName string) error {
	if err := common.PropertyDelete("scheduler-k3s", "--global", getNodeSysctlsProperty(profileName)); err != nil {
		return fmt.Errorf("Unable to delete node sysctls: %w", err)
	}

	if err := isKubernetesAvailable(); err != nil {
		common.LogDebug("kubernetes not available, skipping node sysctls deletion")
		return nil
	}

	return deleteNodeSysctlsRelease(getNodeSysctlsReleaseName(profileName))
}

// nodeSysctlsScopeLabel returns a human readable name for a node sysctls scope
func nodeSysctlsScopeLabel(profileName string) string {
	if profileName == "" {
		return "unprofiled nodes"
	}

	return fmt.Sprintf("node profile %s", profileName)
}
