package scheduler_k3s

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

const (
	annotationsReportType     = "scheduler-k3s-annotations"
	labelsReportType          = "scheduler-k3s-labels"
	autoscalingAuthReportType = "scheduler-k3s-autoscaling-auth"

	// reportProcessTypeGlobal is the rendered form of GlobalProcessType in annotation
	// and label report keys (both JSON and single-flag). Real process types cannot
	// collide because they cannot contain "--".
	reportProcessTypeGlobal = "global"
)

// tokenMask is shown in place of the raw token value in default stdout output.
const tokenMask = "*******"

// ReportSingleApp is an internal function that displays the scheduler-k3s report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	var flags map[string]common.ReportFunc
	if appName == "--global" {
		tokenFlag := "--scheduler-k3s-global-token"
		tokenReportFunc := reportMaskedGlobalToken
		if format == "json" || infoFlag == tokenFlag {
			tokenReportFunc = reportGlobalToken
		}

		flags = map[string]common.ReportFunc{
			"--scheduler-k3s-computed-deploy-timeout":         reportComputedDeployTimeout,
			"--scheduler-k3s-global-deploy-timeout":           reportGlobalDeployTimeout,
			"--scheduler-k3s-computed-image-pull-secrets":     reportComputedImagePullSecrets,
			"--scheduler-k3s-global-image-pull-secrets":       reportGlobalImagePullSecrets,
			"--scheduler-k3s-computed-ingress-class":          reportComputedIngressClass,
			"--scheduler-k3s-global-ingress-class":            reportGlobalIngressClass,
			"--scheduler-k3s-computed-kubeconfig-path":        reportComputedKubeconfigPath,
			"--scheduler-k3s-global-kubeconfig-path":          reportGlobalKubeconfigPath,
			"--scheduler-k3s-computed-kube-context":           reportComputedKubeContext,
			"--scheduler-k3s-global-kube-context":             reportGlobalKubeContext,
			"--scheduler-k3s-computed-kustomize-root-path":    reportComputedKustomizeRootPath,
			"--scheduler-k3s-global-kustomize-root-path":      reportGlobalKustomizeRootPath,
			"--scheduler-k3s-computed-letsencrypt-server":     reportComputedLetsencryptServer,
			"--scheduler-k3s-global-letsencrypt-server":       reportGlobalLetsencryptServer,
			"--scheduler-k3s-computed-letsencrypt-email-prod": reportComputedLetsencryptEmailProd,
			"--scheduler-k3s-global-letsencrypt-email-prod":   reportGlobalLetsencryptEmailProd,
			"--scheduler-k3s-computed-letsencrypt-email-stag": reportComputedLetsencryptEmailStag,
			"--scheduler-k3s-global-letsencrypt-email-stag":   reportGlobalLetsencryptEmailStag,
			"--scheduler-k3s-computed-namespace":              reportComputedNamespace,
			"--scheduler-k3s-global-namespace":                reportGlobalNamespace,
			"--scheduler-k3s-computed-network-interface":      reportComputedNetworkInterface,
			"--scheduler-k3s-global-network-interface":        reportGlobalNetworkInterface,
			"--scheduler-k3s-computed-rollback-on-failure":    reportComputedRollbackOnFailure,
			"--scheduler-k3s-global-rollback-on-failure":      reportGlobalRollbackOnFailure,
			"--scheduler-k3s-computed-shm-size":               reportComputedShmSize,
			"--scheduler-k3s-global-shm-size":                 reportGlobalShmSize,
			tokenFlag:                                         tokenReportFunc,
		}
	} else {
		flags = map[string]common.ReportFunc{
			"--scheduler-k3s-computed-deploy-timeout":         reportComputedDeployTimeout,
			"--scheduler-k3s-deploy-timeout":                  reportDeployTimeout,
			"--scheduler-k3s-global-deploy-timeout":           reportGlobalDeployTimeout,
			"--scheduler-k3s-computed-image-pull-secrets":     reportComputedImagePullSecrets,
			"--scheduler-k3s-image-pull-secrets":              reportImagePullSecrets,
			"--scheduler-k3s-global-image-pull-secrets":       reportGlobalImagePullSecrets,
			"--scheduler-k3s-computed-ingress-class":          reportComputedIngressClass,
			"--scheduler-k3s-global-ingress-class":            reportGlobalIngressClass,
			"--scheduler-k3s-computed-kubeconfig-path":        reportComputedKubeconfigPath,
			"--scheduler-k3s-global-kubeconfig-path":          reportGlobalKubeconfigPath,
			"--scheduler-k3s-computed-kube-context":           reportComputedKubeContext,
			"--scheduler-k3s-global-kube-context":             reportGlobalKubeContext,
			"--scheduler-k3s-computed-kustomize-root-path":    reportComputedKustomizeRootPath,
			"--scheduler-k3s-kustomize-root-path":             reportKustomizeRootPath,
			"--scheduler-k3s-global-kustomize-root-path":      reportGlobalKustomizeRootPath,
			"--scheduler-k3s-computed-letsencrypt-server":     reportComputedLetsencryptServer,
			"--scheduler-k3s-letsencrypt-server":              reportLetsencryptServer,
			"--scheduler-k3s-global-letsencrypt-server":       reportGlobalLetsencryptServer,
			"--scheduler-k3s-computed-letsencrypt-email-prod": reportComputedLetsencryptEmailProd,
			"--scheduler-k3s-global-letsencrypt-email-prod":   reportGlobalLetsencryptEmailProd,
			"--scheduler-k3s-computed-letsencrypt-email-stag": reportComputedLetsencryptEmailStag,
			"--scheduler-k3s-global-letsencrypt-email-stag":   reportGlobalLetsencryptEmailStag,
			"--scheduler-k3s-computed-namespace":              reportComputedNamespace,
			"--scheduler-k3s-namespace":                       reportNamespace,
			"--scheduler-k3s-global-namespace":                reportGlobalNamespace,
			"--scheduler-k3s-computed-network-interface":      reportComputedNetworkInterface,
			"--scheduler-k3s-global-network-interface":        reportGlobalNetworkInterface,
			"--scheduler-k3s-computed-rollback-on-failure":    reportComputedRollbackOnFailure,
			"--scheduler-k3s-rollback-on-failure":             reportRollbackOnFailure,
			"--scheduler-k3s-global-rollback-on-failure":      reportGlobalRollbackOnFailure,
			"--scheduler-k3s-computed-shm-size":               reportComputedShmSize,
			"--scheduler-k3s-global-shm-size":                 reportGlobalShmSize,
			"--scheduler-k3s-shm-size":                        reportShmSize,
		}
	}

	for _, chart := range HelmCharts {
		chartOverrides, err := common.PropertyMapGet("scheduler-k3s", "--global", "chart-overrides."+chart.ReleaseName)
		if err != nil {
			return fmt.Errorf("Unable to get property list: %w", err)
		}
		for key, value := range chartOverrides {
			flagName := fmt.Sprintf("--scheduler-k3s-global-chart.%s.%s", chart.ReleaseName, key)
			value := value
			flags[flagName] = func(appName string) string {
				return value
			}
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "scheduler-k3s",
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

// autoscalingAuthReportEntry is a parsed trigger authentication metadata entry.
type autoscalingAuthReportEntry struct {
	Trigger     string
	MetadataKey string
	Value       string
}

// ReportAutoscalingAuthSingleApp displays the scheduler-k3s autoscaling-auth report for one app.
func ReportAutoscalingAuthSingleApp(appName string, format string, includeMetadata bool, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	entries, err := collectAutoscalingAuthEntries(appName)
	if err != nil {
		return err
	}

	return renderAutoscalingAuthReport(autoscalingAuthRenderInput{
		AppName:         appName,
		Entries:         entries,
		Format:          format,
		IncludeMetadata: includeMetadata,
		InfoFlag:        infoFlag,
	})
}

// collectAutoscalingAuthEntries scans the property store for trigger authentication metadata.
func collectAutoscalingAuthEntries(appName string) ([]autoscalingAuthReportEntry, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, TriggerAuthPropertyPrefix)
	if err != nil {
		return nil, fmt.Errorf("Unable to get property list: %w", err)
	}

	entries := []autoscalingAuthReportEntry{}
	for key, value := range properties {
		metadataKey := strings.TrimPrefix(key, TriggerAuthPropertyPrefix)
		parts := strings.SplitN(metadataKey, ".", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}

		entries = append(entries, autoscalingAuthReportEntry{
			Trigger:     parts[0],
			MetadataKey: parts[1],
			Value:       value,
		})
	}

	return entries, nil
}

type autoscalingAuthRenderInput struct {
	AppName         string
	Entries         []autoscalingAuthReportEntry
	Format          string
	IncludeMetadata bool
	InfoFlag        string
}

// renderAutoscalingAuthReport dispatches to stdout/json/single-flag rendering for trigger auth.
func renderAutoscalingAuthReport(input autoscalingAuthRenderInput) error {
	if input.Format != "stdout" && input.Format != "json" {
		return fmt.Errorf("Invalid format: %s", input.Format)
	}
	if input.Format == "json" && input.InfoFlag != "" {
		return fmt.Errorf("--format flag cannot be specified when specifying an info flag")
	}

	flatValues := map[string]string{}
	flagToValue := map[string]string{}
	flagPrefix := "--" + autoscalingAuthReportType + "."
	for _, entry := range input.Entries {
		composedKey := entry.Trigger + "." + entry.MetadataKey
		flatValues[composedKey] = entry.Value
		flagToValue[flagPrefix+composedKey] = entry.Value
	}

	if input.InfoFlag != "" {
		value, ok := flagToValue[input.InfoFlag]
		if !ok {
			validFlags := []string{}
			for flag := range flagToValue {
				validFlags = append(validFlags, flag)
			}
			sort.Strings(validFlags)
			return fmt.Errorf("Invalid flag passed, valid flags: %s", strings.Join(validFlags, ", "))
		}
		if value != "" {
			fmt.Println(tokenMask)
		} else {
			fmt.Println("")
		}
		return nil
	}

	if input.Format == "json" {
		b, err := json.Marshal(flatValues)
		if err != nil {
			return fmt.Errorf("Unable to marshal json: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	triggers := map[string]bool{}
	for _, entry := range input.Entries {
		triggers[entry.Trigger] = true
	}

	triggerKeys := []string{}
	for trigger := range triggers {
		triggerKeys = append(triggerKeys, trigger)
	}
	sort.Strings(triggerKeys)

	common.LogInfo2Quiet(fmt.Sprintf("%s autoscaling-auth information", input.AppName))
	if len(triggerKeys) == 0 {
		return nil
	}

	length := 31
	if input.IncludeMetadata {
		for _, entry := range input.Entries {
			label := fmt.Sprintf("%s %s:", common.UcFirst(entry.Trigger), entry.MetadataKey)
			if len(label) > length {
				length = len(label)
			}
		}
	}

	for _, trigger := range triggerKeys {
		label := fmt.Sprintf("%s:", common.UcFirst(trigger))
		common.LogVerbose(fmt.Sprintf("%s%s", common.RightPad(label, length, " "), "configured"))

		if !input.IncludeMetadata {
			continue
		}

		triggerEntries := []autoscalingAuthReportEntry{}
		for _, entry := range input.Entries {
			if entry.Trigger == trigger {
				triggerEntries = append(triggerEntries, entry)
			}
		}
		sort.Slice(triggerEntries, func(i, j int) bool {
			return triggerEntries[i].MetadataKey < triggerEntries[j].MetadataKey
		})

		for _, entry := range triggerEntries {
			label := fmt.Sprintf("%s %s:", common.UcFirst(entry.Trigger), entry.MetadataKey)
			common.LogVerbose(fmt.Sprintf("%s%s", common.RightPad(label, length, " "), entry.Value))
		}
	}

	return nil
}

// ExtractReportInfoFlag scans arguments for a single info flag whose name starts with
// the given prefix (e.g. "--scheduler-k3s-annotations.") and returns it separately from
// the remaining passthrough arguments. Returns an error if more than one matching flag
// is present. Matching args are removed from the passthrough list so the caller's
// flag.FlagSet sees only its own flags and positional arguments.
func ExtractReportInfoFlag(prefix string, arguments []string) ([]string, string, error) {
	passthrough := []string{}
	infoFlag := ""
	for _, argument := range arguments {
		if strings.HasPrefix(argument, prefix) {
			if infoFlag != "" {
				return nil, "", fmt.Errorf("only a single info flag may be specified")
			}
			infoFlag = argument
			continue
		}
		passthrough = append(passthrough, argument)
	}
	return passthrough, infoFlag, nil
}

// annotationsLabelsReportEntry is a parsed (processType, resourceType, key) triple
// produced when scanning the property store for annotations or labels.
type annotationsLabelsReportEntry struct {
	ProcessType  string
	ResourceType string
	Key          string
	Value        string
}

// ReportAnnotationsSingleApp displays the scheduler-k3s annotations report for one app.
func ReportAnnotationsSingleApp(appName string, format string, processType string, resourceType string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	entries, err := collectAnnotationsEntries(appName, processType, resourceType)
	if err != nil {
		return err
	}

	return renderAnnotationsLabelsReport(annotationsLabelsRenderInput{
		AppName:    appName,
		ReportType: annotationsReportType,
		Heading:    "annotations",
		RowLabel:   "Annotation",
		Entries:    entries,
		Format:     format,
		InfoFlag:   infoFlag,
	})
}

// ReportLabelsSingleApp displays the scheduler-k3s labels report for one app.
func ReportLabelsSingleApp(appName string, format string, processType string, resourceType string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	entries, err := collectLabelsEntries(appName, processType, resourceType)
	if err != nil {
		return err
	}

	return renderAnnotationsLabelsReport(annotationsLabelsRenderInput{
		AppName:    appName,
		ReportType: labelsReportType,
		Heading:    "labels",
		RowLabel:   "Label",
		Entries:    entries,
		Format:     format,
		InfoFlag:   infoFlag,
	})
}

// collectAnnotationsEntries scans the property store for annotation entries on appName,
// optionally filtered by processType/resourceType.
func collectAnnotationsEntries(appName string, processType string, resourceType string) ([]annotationsLabelsReportEntry, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, "")
	if err != nil {
		return nil, fmt.Errorf("Unable to get property list: %w", err)
	}

	knownResourceTypes := map[string]bool{}
	for _, rt := range AnnotationResourceTypes {
		knownResourceTypes[rt] = true
	}

	entries := []annotationsLabelsReportEntry{}
	for propertyName := range properties {
		if isReservedAnnotationProperty(propertyName) {
			continue
		}

		dot := strings.LastIndex(propertyName, ".")
		if dot <= 0 || dot == len(propertyName)-1 {
			continue
		}

		propProcessType := propertyName[:dot]
		propResourceType := propertyName[dot+1:]
		if !knownResourceTypes[propResourceType] {
			continue
		}

		if processType != "" && propProcessType != processType {
			continue
		}
		if resourceType != "" && propResourceType != resourceType {
			continue
		}

		annotationMap, err := getAnnotation(appName, propProcessType, propResourceType)
		if err != nil {
			return nil, fmt.Errorf("Unable to read annotation %s: %w", propertyName, err)
		}

		for key, value := range annotationMap {
			entries = append(entries, annotationsLabelsReportEntry{
				ProcessType:  propProcessType,
				ResourceType: propResourceType,
				Key:          key,
				Value:        value,
			})
		}
	}

	return entries, nil
}

// collectLabelsEntries scans the property store for label entries on appName,
// optionally filtered by processType/resourceType.
func collectLabelsEntries(appName string, processType string, resourceType string) ([]annotationsLabelsReportEntry, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, "labels.")
	if err != nil {
		return nil, fmt.Errorf("Unable to get property list: %w", err)
	}

	knownResourceTypes := map[string]bool{}
	for _, rt := range LabelResourceTypes {
		knownResourceTypes[rt] = true
	}

	entries := []annotationsLabelsReportEntry{}
	for propertyName := range properties {
		suffix := strings.TrimPrefix(propertyName, "labels.")
		dot := strings.LastIndex(suffix, ".")
		if dot <= 0 || dot == len(suffix)-1 {
			continue
		}

		propProcessType := suffix[:dot]
		propResourceType := suffix[dot+1:]
		if !knownResourceTypes[propResourceType] {
			continue
		}

		if processType != "" && propProcessType != processType {
			continue
		}
		if resourceType != "" && propResourceType != resourceType {
			continue
		}

		labelMap, err := getLabel(appName, propProcessType, propResourceType)
		if err != nil {
			return nil, fmt.Errorf("Unable to read label %s: %w", propertyName, err)
		}

		for key, value := range labelMap {
			entries = append(entries, annotationsLabelsReportEntry{
				ProcessType:  propProcessType,
				ResourceType: propResourceType,
				Key:          key,
				Value:        value,
			})
		}
	}

	return entries, nil
}

func isReservedAnnotationProperty(propertyName string) bool {
	for _, prefix := range reservedAnnotationPrefixes {
		if strings.HasPrefix(propertyName, prefix) {
			return true
		}
	}
	return false
}

type annotationsLabelsRenderInput struct {
	AppName    string
	ReportType string
	// Heading is the plural noun used in stdout headers (e.g. "annotations", "labels").
	Heading string
	// RowLabel is the singular noun used to prefix each stdout row (e.g. "Annotation").
	RowLabel string
	Entries  []annotationsLabelsReportEntry
	Format   string
	InfoFlag string
}

// renderAnnotationsLabelsReport dispatches to stdout/json/single-flag rendering for the
// annotations and labels reports.
func renderAnnotationsLabelsReport(input annotationsLabelsRenderInput) error {
	if input.Format != "stdout" && input.Format != "json" {
		return fmt.Errorf("Invalid format: %s", input.Format)
	}
	if input.Format == "json" && input.InfoFlag != "" {
		return fmt.Errorf("--format flag cannot be specified when specifying an info flag")
	}

	flatValues := map[string]string{}
	flagToValue := map[string]string{}
	flagPrefix := "--" + input.ReportType + "."
	for _, entry := range input.Entries {
		composedKey := renderProcessType(entry.ProcessType) + "." + entry.ResourceType + "." + entry.Key
		flatValues[composedKey] = entry.Value
		flagToValue[flagPrefix+composedKey] = entry.Value
	}

	if input.InfoFlag != "" {
		value, ok := flagToValue[input.InfoFlag]
		if !ok {
			validFlags := []string{}
			for flag := range flagToValue {
				validFlags = append(validFlags, flag)
			}
			sort.Strings(validFlags)
			return fmt.Errorf("Invalid flag passed, valid flags: %s", strings.Join(validFlags, ", "))
		}
		fmt.Println(value)
		return nil
	}

	if input.Format == "json" {
		b, err := json.Marshal(flatValues)
		if err != nil {
			return fmt.Errorf("Unable to marshal json: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	entriesByGroup := map[string][]annotationsLabelsReportEntry{}
	for _, entry := range input.Entries {
		groupKey := renderProcessType(entry.ProcessType) + "/" + entry.ResourceType
		entriesByGroup[groupKey] = append(entriesByGroup[groupKey], entry)
	}

	groupKeys := []string{}
	for groupKey := range entriesByGroup {
		groupKeys = append(groupKeys, groupKey)
	}
	sort.Strings(groupKeys)

	common.LogInfo2Quiet(fmt.Sprintf("%s %s information", input.AppName, input.Heading))
	if len(groupKeys) == 0 {
		return nil
	}

	for _, groupKey := range groupKeys {
		groupEntries := entriesByGroup[groupKey]
		sort.Slice(groupEntries, func(i, j int) bool {
			return groupEntries[i].Key < groupEntries[j].Key
		})

		length := 31
		for _, entry := range groupEntries {
			label := fmt.Sprintf("%s (%s) %s:", input.RowLabel, groupKey, entry.Key)
			if len(label) > length {
				length = len(label)
			}
		}

		for _, entry := range groupEntries {
			label := fmt.Sprintf("%s (%s) %s:", input.RowLabel, groupKey, entry.Key)
			common.LogVerbose(fmt.Sprintf("%s%s", common.RightPad(label, length, " "), entry.Value))
		}
	}

	return nil
}

// renderProcessType normalizes the in-storage process type for display in JSON keys,
// stdout headers, and single-flag lookups. GlobalProcessType (literal "--global") is
// rendered as "global"; real process types pass through unchanged.
func renderProcessType(processType string) string {
	if processType == GlobalProcessType {
		return reportProcessTypeGlobal
	}
	return processType
}

func reportComputedDeployTimeout(appName string) string {
	return getComputedDeployTimeout(appName)
}

func reportDeployTimeout(appName string) string {
	return getDeployTimeout(appName)
}

func reportGlobalDeployTimeout(appName string) string {
	return getGlobalDeployTimeout()
}

func reportComputedImagePullSecrets(appName string) string {
	return getComputedImagePullSecrets(appName)
}

func reportImagePullSecrets(appName string) string {
	return getImagePullSecrets(appName)
}

func reportGlobalImagePullSecrets(appName string) string {
	return getGlobalImagePullSecrets()
}

func reportComputedIngressClass(appName string) string {
	return getComputedIngressClass()
}

func reportGlobalIngressClass(appName string) string {
	return getGlobalIngressClass()
}

func reportComputedKubeconfigPath(appName string) string {
	return getComputedKubeconfigPath()
}

func reportGlobalKubeconfigPath(appName string) string {
	return getGlobalKubeconfigPath()
}

func reportComputedKubeContext(appName string) string {
	return getComputedKubeContext()
}

func reportGlobalKubeContext(appName string) string {
	return getGlobalKubeContext()
}

func reportComputedLetsencryptServer(appName string) string {
	return getComputedLetsencryptServer(appName)
}

func reportLetsencryptServer(appName string) string {
	return getLetsencryptServer(appName)
}

func reportGlobalLetsencryptServer(appName string) string {
	return getGlobalLetsencryptServer()
}

func reportComputedLetsencryptEmailProd(appName string) string {
	return getComputedLetsencryptEmailProd()
}

func reportGlobalLetsencryptEmailProd(appName string) string {
	return getGlobalLetsencryptEmailProd()
}

func reportComputedLetsencryptEmailStag(appName string) string {
	return getComputedLetsencryptEmailStag()
}

func reportGlobalLetsencryptEmailStag(appName string) string {
	return getGlobalLetsencryptEmailStag()
}

func reportComputedKustomizeRootPath(appName string) string {
	return getComputedKustomizeRootPath(appName)
}

func reportKustomizeRootPath(appName string) string {
	return getKustomizeRootPath(appName)
}

func reportGlobalKustomizeRootPath(appName string) string {
	return getGlobalKustomizeRootPath()
}

func reportComputedNamespace(appName string) string {
	return getComputedNamespace(appName)
}

func reportNamespace(appName string) string {
	return getNamespace(appName)
}

func reportGlobalNamespace(appName string) string {
	return getGlobalNamespace()
}

func reportComputedNetworkInterface(appName string) string {
	return getComputedNetworkInterface()
}

func reportGlobalNetworkInterface(appName string) string {
	return getGlobalNetworkInterface()
}

func reportComputedRollbackOnFailure(appName string) string {
	return getComputedRollbackOnFailure(appName)
}

func reportRollbackOnFailure(appName string) string {
	return getRollbackOnFailure(appName)
}

func reportGlobalRollbackOnFailure(appName string) string {
	return getGlobalRollbackOnFailure()
}

func reportComputedShmSize(appName string) string {
	return getComputedShmSize(appName)
}

func reportGlobalShmSize(appName string) string {
	return getGlobalShmSize()
}

func reportShmSize(appName string) string {
	return getShmSize(appName)
}

func reportGlobalToken(appName string) string {
	return getGlobalGlobalToken()
}

func reportMaskedGlobalToken(appName string) string {
	value := getGlobalGlobalToken()
	if value == "" {
		return ""
	}
	return tokenMask
}
