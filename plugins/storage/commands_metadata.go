package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// entryMapField selects which of an Entry's string maps a metadata command
// operates on, so the annotations and labels commands can share one
// implementation the way scheduler-k3s does.
type entryMapField struct {
	Name       string
	Singular   string
	ReportType string
	RowLabel   string
	Get        func(*Entry) map[string]string
	Set        func(*Entry, map[string]string)
}

var annotationsField = entryMapField{
	Name:       "annotations",
	Singular:   "annotation",
	ReportType: "storage-annotations",
	RowLabel:   "Annotation",
	Get:        func(e *Entry) map[string]string { return e.Annotations },
	Set:        func(e *Entry, m map[string]string) { e.Annotations = m },
}

var labelsField = entryMapField{
	Name:       "labels",
	Singular:   "label",
	ReportType: "storage-labels",
	RowLabel:   "Label",
	Get:        func(e *Entry) map[string]string { return e.Labels },
	Set:        func(e *Entry, m map[string]string) { e.Labels = m },
}

// CommandAnnotationsSet sets or clears a single annotation on an entry.
func CommandAnnotationsSet(name string, key string, value string) error {
	return setEntryMapKey(annotationsField, name, key, value)
}

// CommandLabelsSet sets or clears a single label on an entry.
func CommandLabelsSet(name string, key string, value string) error {
	return setEntryMapKey(labelsField, name, key, value)
}

// CommandAnnotationsReport displays the annotations on one entry, or on
// every entry when no name is given.
func CommandAnnotationsReport(name string, format string, infoFlag string) error {
	return reportEntryMap(annotationsField, name, format, infoFlag)
}

// CommandLabelsReport displays the labels on one entry, or on every entry
// when no name is given.
func CommandLabelsReport(name string, format string, infoFlag string) error {
	return reportEntryMap(labelsField, name, format, infoFlag)
}

// setEntryMapKey writes a single key on one of an entry's metadata maps.
// An empty value deletes just that key, leaving its siblings in place -
// the behavior the wholesale --annotation flag could never express.
func setEntryMapKey(field entryMapField, name string, key string, value string) error {
	if name == "" {
		return errors.New("storage entry name is required")
	}
	if key == "" {
		return fmt.Errorf("No %s key specified", field.Singular)
	}
	if !EntryExists(name) {
		return fmt.Errorf("storage entry %q does not exist", name)
	}

	entry, err := LoadEntry(name)
	if err != nil {
		return err
	}

	values := field.Get(entry)
	if value == "" {
		delete(values, key)
		common.LogInfo2Quiet(fmt.Sprintf("Unsetting %s %s", field.Singular, key))
	} else {
		if values == nil {
			values = map[string]string{}
		}
		values[key] = value
		common.LogInfo2Quiet(fmt.Sprintf("Setting %s %s to %s", field.Singular, key, value))
	}
	if len(values) == 0 {
		// Drop the empty map so the omitempty tag keeps it out of the JSON.
		values = nil
	}
	field.Set(entry, values)

	if err := entry.Validate(); err != nil {
		return err
	}
	if err := SaveEntry(entry); err != nil {
		return err
	}

	// k3s renders annotations and labels onto the PVC and PV through the
	// entry's helm release, so the cluster only sees this once the chart
	// is re-applied.
	if entry.Scheduler == SchedulerK3s {
		if err := callSchedulerCreateTrigger(entry); err != nil {
			return fmt.Errorf("scheduler refused %s change for %q: %w", field.Name, name, err)
		}
	}
	return nil
}

// reportEntryMap renders one entry's metadata map, or every entry's when
// no name is given.
func reportEntryMap(field entryMapField, name string, format string, infoFlag string) error {
	if format != "stdout" && format != "text" && format != "json" {
		return fmt.Errorf("Invalid format: %s", format)
	}
	if format == "json" && infoFlag != "" {
		return errors.New("--format flag cannot be specified when specifying an info flag")
	}

	if name != "" {
		if !EntryExists(name) {
			return fmt.Errorf("storage entry %q does not exist", name)
		}
		entry, err := LoadEntry(name)
		if err != nil {
			return err
		}
		return renderEntryMap(field, entry, format, infoFlag)
	}

	if infoFlag != "" {
		return fmt.Errorf("storage:%s:report requires a storage entry name when an info flag is specified", field.Name)
	}

	entries, err := ListEntries()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		common.LogInfo1Quiet("No storage entries registered")
		return nil
	}
	for _, entry := range entries {
		if err := renderEntryMap(field, entry, format, ""); err != nil {
			return err
		}
	}
	return nil
}

// renderEntryMap prints one entry's metadata map. Keys are emitted
// verbatim rather than through common.ReportSingleApp, which rewrites
// dots and dashes into spaces and would mangle keys like
// app.kubernetes.io/part-of.
func renderEntryMap(field entryMapField, entry *Entry, format string, infoFlag string) error {
	values := field.Get(entry)

	keys := []string{}
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	if infoFlag != "" {
		flagPrefix := "--" + field.ReportType + "."
		validFlags := []string{}
		for _, key := range keys {
			flag := flagPrefix + key
			if flag == infoFlag {
				fmt.Println(values[key])
				return nil
			}
			validFlags = append(validFlags, flag)
		}
		return fmt.Errorf("Invalid flag passed, valid flags: %s", strings.Join(validFlags, ", "))
	}

	if format == "json" {
		flat := map[string]string{}
		for _, key := range keys {
			flat[key] = values[key]
		}
		data, err := json.Marshal(flat)
		if err != nil {
			return fmt.Errorf("Unable to marshal json: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	common.LogInfo2Quiet(fmt.Sprintf("%s %s information", entry.Name, field.Name))
	length := 31
	for _, key := range keys {
		if label := fmt.Sprintf("%s %s:", field.RowLabel, key); len(label) > length {
			length = len(label)
		}
	}
	for _, key := range keys {
		label := fmt.Sprintf("%s %s:", field.RowLabel, key)
		common.LogVerbose(fmt.Sprintf("%s%s", common.RightPad(label, length, " "), values[key]))
	}
	return nil
}
