package dockeroptions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const dockerOptionsReportType = "docker-options"

// ReportSingleApp displays the docker options report for a single app.
// Default-scope options are reported under fixed keys for each phase.
// Process-scoped options surface as dynamic per-process keys, one per
// configured process+phase combination.
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		flags = map[string]common.ReportFunc{}
	} else {
		flags = map[string]common.ReportFunc{
			"--docker-options-build":  reportBuildOptions,
			"--docker-options-deploy": reportDeployOptions,
			"--docker-options-run":    reportRunOptions,
		}
	}

	processTypes, err := ListProcessTypesWithOptions(appName)
	if err != nil {
		return err
	}
	for _, processType := range processTypes {
		processType := processType
		flagName := fmt.Sprintf("--docker-options-deploy.%s", processType)
		flags[flagName] = func(app string) string {
			return joinProcessPhaseOptions(app, processType, "deploy")
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)

	if format == "" {
		format = "stdout"
	}
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Format must be \"stdout\" or \"json\": %q", format)
	}
	if format == "json" && infoFlag != "" {
		return fmt.Errorf("--format flag cannot be specified when specifying an info flag")
	}

	if format == "json" {
		data, err := buildJSONReportData(appName, infoFlags)
		if err != nil {
			return err
		}
		out, err := json.Marshal(data)
		if err != nil {
			return err
		}
		common.Log(string(out))
		return nil
	}

	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              dockerOptionsReportType,
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        true,
	})
}

// listReportKey returns the JSON key used for the structured list companion to
// a shorthand report key (e.g. "deploy" -> "deploy-list").
func listReportKey(shorthandKey string) string {
	return shorthandKey + "-list"
}

// buildJSONReportData assembles the docker-options JSON report. Existing string
// keys match common.ReportSingleApp output; shorthand keys also gain parallel
// -list keys whose values are the stored option slices.
func buildJSONReportData(appName string, infoFlags map[string]string) (map[string]any, error) {
	data := map[string]any{}
	pluginPrefix := fmt.Sprintf("--%s-", dockerOptionsReportType)
	for key, value := range infoFlags {
		legacyKey := strings.TrimPrefix(key, "--")
		if strings.HasPrefix(key, pluginPrefix) {
			data[strings.TrimPrefix(key, pluginPrefix)] = value
			data[legacyKey] = value
		} else {
			data[legacyKey] = value
		}
	}

	if appName == "--global" {
		return data, nil
	}

	phases := []struct {
		phase        string
		shorthandKey string
	}{
		{"build", "build"},
		{"deploy", "deploy"},
		{"run", "run"},
	}
	for _, phase := range phases {
		options, err := GetDockerOptionsForProcessPhase(appName, DefaultProcessType, phase.phase)
		if err != nil {
			return nil, err
		}
		data[listReportKey(phase.shorthandKey)] = options
	}

	processTypes, err := ListProcessTypesWithOptions(appName)
	if err != nil {
		return nil, err
	}
	for _, processType := range processTypes {
		options, err := GetDockerOptionsForProcessPhase(appName, processType, "deploy")
		if err != nil {
			return nil, err
		}
		data[listReportKey(fmt.Sprintf("deploy.%s", processType))] = options
	}

	return data, nil
}

func reportBuildOptions(appName string) string {
	return joinProcessPhaseOptions(appName, DefaultProcessType, "build")
}

func reportDeployOptions(appName string) string {
	return joinProcessPhaseOptions(appName, DefaultProcessType, "deploy")
}

func reportRunOptions(appName string) string {
	return joinProcessPhaseOptions(appName, DefaultProcessType, "run")
}

func joinProcessPhaseOptions(appName, processType, phase string) string {
	options, err := GetDockerOptionsForProcessPhase(appName, processType, phase)
	if err != nil || len(options) == 0 {
		return ""
	}
	return strings.Join(options, " ")
}
