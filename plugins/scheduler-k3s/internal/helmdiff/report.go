// Derived from github.com/databus23/helm-diff (Apache License 2.0).
// See LICENSE and NOTICE.md in this directory.

package helmdiff

import (
	"fmt"
	"io"

	"github.com/aryann/difflib"
	"github.com/mgutz/ansi"
)

// Report stores diff report data and formatting state.
type Report struct {
	format      ReportFormat
	Entries     []ReportEntry
	mode        string
	findRenames float32
}

// ReportEntry stores changes for a single Kubernetes object.
type ReportEntry struct {
	Key             string
	SuppressedKinds []string
	Kind            string
	Context         int
	Diffs           []difflib.DiffRecord
	ChangeType      string
}

// ReportFormat holds the output renderer and change-type styles.
type ReportFormat struct {
	output       func(r *Report, to io.Writer)
	changestyles map[string]ChangeStyle
}

// ChangeStyle controls how a single change type is rendered.
type ChangeStyle struct {
	color   string
	message string
}

// setupReportFormat configures the report's output renderer.
// Only the default unified-diff format is supported.
func (r *Report) setupReportFormat(format string) {
	r.mode = format
	setupDiffReport(r)
}

// addEntry appends a single diff entry to the report.
func (r *Report) addEntry(key string, suppressedKinds []string, kind string, context int, diffs []difflib.DiffRecord, changeType string) {
	entry := ReportEntry{
		key,
		suppressedKinds,
		kind,
		context,
		diffs,
		changeType,
	}
	r.Entries = append(r.Entries, entry)
}

// print writes all entries via the configured output renderer.
func (r *Report) print(to io.Writer) {
	r.format.output(r, to)
}

// clean resets the entries slice (used by tests).
func (r *Report) clean() {
	r.Entries = nil
}

func setupDiffReport(r *Report) {
	r.format.output = printDiffReport
	r.format.changestyles = make(map[string]ChangeStyle)
	r.format.changestyles["ADD"] = ChangeStyle{color: "green", message: "has been added:"}
	r.format.changestyles["REMOVE"] = ChangeStyle{color: "red", message: "has been removed:"}
	r.format.changestyles["MODIFY"] = ChangeStyle{color: "yellow", message: "has changed:"}
	r.format.changestyles["MODIFY_SUPPRESSED"] = ChangeStyle{color: "blue+h", message: "has changed, but diff is empty after suppression."}
}

func printDiffReport(r *Report, to io.Writer) {
	for _, entry := range r.Entries {
		_, _ = fmt.Fprintf(
			to,
			ansi.Color("%s %s", r.format.changestyles[entry.ChangeType].color)+"\n",
			entry.Key,
			r.format.changestyles[entry.ChangeType].message,
		)
		printDiffRecords(entry.SuppressedKinds, entry.Kind, entry.Context, entry.Diffs, to)
	}
}
