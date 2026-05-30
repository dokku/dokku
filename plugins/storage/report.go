package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func ReportSingleApp(appName string, format string, infoFlag string) error {
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
	}

	flags := map[string]common.ReportFunc{}
	if appName != "--global" {
		flags["--storage-build-mounts"] = reportBuildMounts
		flags["--storage-deploy-mounts"] = reportDeployMounts
		flags["--storage-run-mounts"] = reportRunMounts

		attachments, err := LoadAttachments(appName)
		if err != nil {
			common.LogWarn(fmt.Sprintf("Unable to load storage attachments for %q: %s", appName, err))
			attachments = nil
		}
		index := 0
		for _, attachment := range attachments {
			entry, err := LoadEntry(attachment.EntryName)
			if err != nil {
				common.LogWarn(fmt.Sprintf("Skipping attachment on %q: missing entry %q (%s)", appName, attachment.EntryName, err))
				continue
			}
			index++
			registerAttachmentFlags(flags, index, attachment, entry)
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "storage",
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

func reportBuildMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "build")
}

func reportDeployMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "deploy")
}

func reportRunMounts(appName string) string {
	return GetBindMountsForDisplay(appName, "run")
}

// registerAttachmentFlags adds one report flag per Attachment field for the
// given 1-based index. Host path is resolved from the referenced Entry
// (falling back to entry.Name when the entry has no host path, matching
// ListAppMountEntries).
func registerAttachmentFlags(flags map[string]common.ReportFunc, index int, att *Attachment, entry *Entry) {
	prefix := fmt.Sprintf("--storage-attachment.%d.", index)
	captured := att
	host := entry.HostPath
	if host == "" {
		host = entry.Name
	}
	flags[prefix+"entry-name"] = func(string) string { return captured.EntryName }
	flags[prefix+"host-path"] = func(string) string { return host }
	flags[prefix+"container-path"] = func(string) string { return captured.ContainerPath }
	flags[prefix+"phases"] = func(string) string { return strings.Join(captured.Phases, ",") }
	flags[prefix+"process-type"] = func(string) string { return captured.ProcessType }
	flags[prefix+"subpath"] = func(string) string { return captured.Subpath }
	flags[prefix+"readonly"] = func(string) string { return strconv.FormatBool(captured.Readonly) }
	flags[prefix+"volume-options"] = func(string) string { return captured.VolumeOptions }
	flags[prefix+"volume-chown"] = func(string) string { return captured.VolumeChown }
}
