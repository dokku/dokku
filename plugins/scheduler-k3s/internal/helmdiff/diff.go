// Derived from github.com/databus23/helm-diff (Apache License 2.0).
// See LICENSE and NOTICE.md in this directory.

package helmdiff

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/aryann/difflib"
	"github.com/mgutz/ansi"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

// Options controls the behavior of Manifests.
type Options struct {
	OutputFormat              string
	OutputContext             int
	StripTrailingCR           bool
	ShowSecrets               bool
	ShowSecretsDecoded        bool
	SuppressedKinds           []string
	FindRenames               float32
	SuppressedOutputLineRegex []string
}

const kindSecret = "Secret"

// Manifests compares two indexed manifest maps and writes a unified diff to to.
// It returns true if any changes were detected.
func Manifests(oldIndex, newIndex map[string]*MappingResult, options *Options, to io.Writer) bool {
	seenAnyChanges, report, err := generateReport(oldIndex, newIndex, options)
	if err != nil {
		panic(err)
	}

	report.print(to)
	report.clean()
	return seenAnyChanges
}

// ManifestReport is like Manifests but returns the constructed Report instead
// of writing it. Useful for tests.
func ManifestReport(oldIndex, newIndex map[string]*MappingResult, options *Options) (*Report, error) {
	_, report, err := generateReport(oldIndex, newIndex, options)
	return report, err
}

func generateReport(oldIndex, newIndex map[string]*MappingResult, options *Options) (bool, *Report, error) {
	report := Report{findRenames: options.FindRenames}
	report.setupReportFormat(options.OutputFormat)
	var possiblyRemoved []string

	for _, key := range sortedKeys(oldIndex) {
		oldContent := oldIndex[key]

		if newContent, ok := newIndex[key]; ok {
			doDiff(&report, key, oldContent, newContent, options)
		} else {
			possiblyRemoved = append(possiblyRemoved, key)
		}
	}

	var possiblyAdded []string
	for _, key := range sortedKeys(newIndex) {
		if _, ok := oldIndex[key]; !ok {
			possiblyAdded = append(possiblyAdded, key)
		}
	}

	removed, added := contentSearch(&report, possiblyRemoved, oldIndex, possiblyAdded, newIndex, options)

	for _, key := range removed {
		oldContent := oldIndex[key]
		if oldContent.ResourcePolicy != "keep" {
			doDiff(&report, key, oldContent, nil, options)
		}
	}

	for _, key := range added {
		newContent := newIndex[key]
		doDiff(&report, key, nil, newContent, options)
	}

	seenAnyChanges := len(report.Entries) > 0

	report, err := doSuppress(report, options.SuppressedOutputLineRegex)

	return seenAnyChanges, &report, err
}

func doSuppress(report Report, suppressedOutputLineRegex []string) (Report, error) {
	if len(report.Entries) == 0 || len(suppressedOutputLineRegex) == 0 {
		return report, nil
	}

	filteredReport := Report{
		findRenames: report.findRenames,
	}
	filteredReport.format = report.format
	filteredReport.Entries = []ReportEntry{}

	var suppressOutputRegexes []*regexp.Regexp

	for _, suppressOutputRegex := range suppressedOutputLineRegex {
		regex, err := regexp.Compile(suppressOutputRegex)
		if err != nil {
			return Report{}, err
		}

		suppressOutputRegexes = append(suppressOutputRegexes, regex)
	}

	for _, entry := range report.Entries {
		var diffs []difflib.DiffRecord

	DIFFS:
		for _, diff := range entry.Diffs {
			for _, suppressOutputRegex := range suppressOutputRegexes {
				if suppressOutputRegex.MatchString(diff.Payload) {
					continue DIFFS
				}
			}

			diffs = append(diffs, diff)
		}

		containsDiff := false

		for _, diff := range diffs {
			if diff.Delta.String() != " " {
				containsDiff = true
				break
			}
		}

		diffRecords := []difflib.DiffRecord{}
		switch {
		case containsDiff:
			diffRecords = diffs
		case entry.ChangeType == "MODIFY":
			entry.ChangeType = "MODIFY_SUPPRESSED"
		}

		filteredReport.addEntry(entry.Key, entry.SuppressedKinds, entry.Kind, entry.Context, diffRecords, entry.ChangeType)
	}

	return filteredReport, nil
}

func actualChanges(diff []difflib.DiffRecord) int {
	changes := 0
	for _, record := range diff {
		if record.Delta != difflib.Common {
			changes++
		}
	}
	return changes
}

const (
	renameDetectionMinLengthRatio float32 = 0.1
	renameDetectionMaxLengthRatio float32 = 10.0
)

func contentSearch(report *Report, possiblyRemoved []string, oldIndex map[string]*MappingResult, possiblyAdded []string, newIndex map[string]*MappingResult, options *Options) ([]string, []string) {
	if options.FindRenames <= 0 {
		return possiblyRemoved, possiblyAdded
	}

	var removed []string

	for _, removedKey := range possiblyRemoved {
		oldContent := oldIndex[removedKey]
		var smallestKey string
		var smallestFraction float32 = math.MaxFloat32
		for _, addedKey := range possiblyAdded {
			newContent := newIndex[addedKey]
			if oldContent.Kind != newContent.Kind {
				continue
			}

			oldLen := len(oldContent.Content)
			newLen := len(newContent.Content)
			if oldLen == 0 || newLen == 0 {
				continue
			}
			if oldContent.Kind != kindSecret {
				ratio := float32(oldLen) / float32(newLen)
				if ratio < renameDetectionMinLengthRatio || ratio > renameDetectionMaxLengthRatio {
					continue
				}
			}

			switch {
			case options.ShowSecretsDecoded:
				decodeSecrets(oldContent, newContent)
			case !options.ShowSecrets:
				redactSecrets(oldContent, newContent)
			}

			diff := diffMappingResults(oldContent, newContent, options.StripTrailingCR)
			delta := actualChanges(diff)
			if delta == 0 || len(diff) == 0 {
				continue
			}
			fraction := float32(delta) / float32(len(diff))
			if fraction > 0 && fraction < smallestFraction {
				smallestKey = addedKey
				smallestFraction = fraction
			}
		}

		if smallestFraction < options.FindRenames {
			index := sort.SearchStrings(possiblyAdded, smallestKey)
			possiblyAdded = append(possiblyAdded[:index], possiblyAdded[index+1:]...)
			newContent := newIndex[smallestKey]
			doDiff(report, removedKey, oldContent, newContent, options)
		} else {
			removed = append(removed, removedKey)
		}
	}

	return removed, possiblyAdded
}

func doDiff(report *Report, key string, oldContent *MappingResult, newContent *MappingResult, options *Options) {
	if oldContent != nil && newContent != nil && oldContent.Content == newContent.Content {
		return
	}
	switch {
	case options.ShowSecretsDecoded:
		decodeSecrets(oldContent, newContent)
	case !options.ShowSecrets:
		redactSecrets(oldContent, newContent)
	}

	var changeType string
	var subjectKind string
	var diffs []difflib.DiffRecord
	switch {
	case oldContent == nil:
		changeType = "ADD"
		if newContent != nil {
			subjectKind = newContent.Kind
			emptyMapping := &MappingResult{}
			diffs = diffMappingResults(emptyMapping, newContent, options.StripTrailingCR)
		}
	case newContent == nil:
		changeType = "REMOVE"
		subjectKind = oldContent.Kind
		emptyMapping := &MappingResult{}
		diffs = diffMappingResults(oldContent, emptyMapping, options.StripTrailingCR)
	default:
		changeType = "MODIFY"
		subjectKind = oldContent.Kind
		diffs = diffMappingResults(oldContent, newContent, options.StripTrailingCR)
		if actualChanges(diffs) == 0 {
			return
		}
	}

	report.addEntry(key, options.SuppressedKinds, subjectKind, options.OutputContext, diffs, changeType)
}

func preHandleSecrets(old, new *MappingResult) (v1.Secret, v1.Secret, error, error) {
	var oldSecretDecodeErr, newSecretDecodeErr error
	var oldSecret, newSecret v1.Secret
	if old != nil {
		oldSecretDecodeErr = yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(old.Content)).Decode(&oldSecret)
		if oldSecretDecodeErr != nil {
			old.Content = fmt.Sprintf("Error parsing old secret: %s", oldSecretDecodeErr)
		} else {
			if len(oldSecret.StringData) > 0 && oldSecret.Data == nil {
				oldSecret.Data = make(map[string][]byte, len(oldSecret.StringData))
			}
			for k, v := range oldSecret.StringData {
				oldSecret.Data[k] = []byte(v)
			}
		}
	}
	if new != nil {
		newSecretDecodeErr = yaml.NewYAMLToJSONDecoder(bytes.NewBufferString(new.Content)).Decode(&newSecret)
		if newSecretDecodeErr != nil {
			new.Content = fmt.Sprintf("Error parsing new secret: %s", newSecretDecodeErr)
		} else {
			if len(newSecret.StringData) > 0 && newSecret.Data == nil {
				newSecret.Data = make(map[string][]byte, len(newSecret.StringData))
			}
			for k, v := range newSecret.StringData {
				newSecret.Data[k] = []byte(v)
			}
		}
	}
	return oldSecret, newSecret, oldSecretDecodeErr, newSecretDecodeErr
}

// redactSecrets replaces Secret data values with placeholders so the actual
// secret bytes never appear in diff output.
func redactSecrets(old, new *MappingResult) {
	if (old != nil && old.Kind != kindSecret) || (new != nil && new.Kind != kindSecret) {
		return
	}
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	oldSecret, newSecret, oldSecretDecodeErr, newSecretDecodeErr := preHandleSecrets(old, new)

	if old != nil && oldSecretDecodeErr == nil {
		oldSecret.StringData = make(map[string]string, len(oldSecret.Data))
		for k, v := range oldSecret.Data {
			if new != nil && bytes.Equal(v, newSecret.Data[k]) {
				oldSecret.StringData[k] = fmt.Sprintf("REDACTED # (%d bytes)", len(v))
			} else {
				oldSecret.StringData[k] = fmt.Sprintf("-------- # (%d bytes)", len(v))
			}
		}
	}
	if new != nil && newSecretDecodeErr == nil {
		newSecret.StringData = make(map[string]string, len(newSecret.Data))
		for k, v := range newSecret.Data {
			if old != nil && bytes.Equal(v, oldSecret.Data[k]) {
				newSecret.StringData[k] = fmt.Sprintf("REDACTED # (%d bytes)", len(v))
			} else {
				newSecret.StringData[k] = fmt.Sprintf("++++++++ # (%d bytes)", len(v))
			}
		}
	}

	if old != nil && oldSecretDecodeErr == nil {
		oldSecretBuf := bytes.NewBuffer(nil)
		oldSecret.Data = nil
		if err := serializer.Encode(&oldSecret, oldSecretBuf); err != nil {
			new.Content = fmt.Sprintf("Error encoding new secret: %s", err)
		}
		old.Content = getComment(old.Content) + strings.Replace(strings.Replace(oldSecretBuf.String(), "stringData", "data", 1), "  creationTimestamp: null\n", "", 1)
		oldSecretBuf.Reset()
	}
	if new != nil && newSecretDecodeErr == nil {
		newSecretBuf := bytes.NewBuffer(nil)
		newSecret.Data = nil
		if err := serializer.Encode(&newSecret, newSecretBuf); err != nil {
			new.Content = fmt.Sprintf("Error encoding new secret: %s", err)
		}
		new.Content = getComment(new.Content) + strings.Replace(strings.Replace(newSecretBuf.String(), "stringData", "data", 1), "  creationTimestamp: null\n", "", 1)
		newSecretBuf.Reset()
	}
}

// decodeSecrets renders Secret data values base64-decoded in diff output.
func decodeSecrets(old, new *MappingResult) {
	if (old != nil && old.Kind != kindSecret) || (new != nil && new.Kind != kindSecret) {
		return
	}
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	oldSecret, newSecret, oldSecretDecodeErr, newSecretDecodeErr := preHandleSecrets(old, new)

	if old != nil && oldSecretDecodeErr == nil {
		oldSecret.StringData = make(map[string]string, len(oldSecret.Data))
		for k, v := range oldSecret.Data {
			oldSecret.StringData[k] = string(v)
		}
	}
	if new != nil && newSecretDecodeErr == nil {
		newSecret.StringData = make(map[string]string, len(newSecret.Data))
		for k, v := range newSecret.Data {
			newSecret.StringData[k] = string(v)
		}
	}

	if old != nil && oldSecretDecodeErr == nil {
		oldSecretBuf := bytes.NewBuffer(nil)
		oldSecret.Data = nil
		if err := serializer.Encode(&oldSecret, oldSecretBuf); err != nil {
			new.Content = fmt.Sprintf("Error encoding new secret: %s", err)
		}
		old.Content = getComment(old.Content) + strings.Replace(oldSecretBuf.String(), "  creationTimestamp: null\n", "", 1)
		oldSecretBuf.Reset()
	}
	if new != nil && newSecretDecodeErr == nil {
		newSecretBuf := bytes.NewBuffer(nil)
		newSecret.Data = nil
		if err := serializer.Encode(&newSecret, newSecretBuf); err != nil {
			new.Content = fmt.Sprintf("Error encoding new secret: %s", err)
		}
		new.Content = getComment(new.Content) + strings.Replace(newSecretBuf.String(), "  creationTimestamp: null\n", "", 1)
		newSecretBuf.Reset()
	}
}

// getComment returns the first line of a string if it is a comment. This
// preserves the leading "# Source: ..." lines from Helm-rendered manifests.
func getComment(s string) string {
	i := strings.Index(s, "\n")
	if i < 0 || !strings.HasPrefix(s, "#") {
		return ""
	}
	return s[:i+1]
}

func diffMappingResults(oldContent *MappingResult, newContent *MappingResult, stripTrailingCR bool) []difflib.DiffRecord {
	return diffStrings(oldContent.Content, newContent.Content, stripTrailingCR)
}

func diffStrings(before, after string, stripTrailingCR bool) []difflib.DiffRecord {
	return difflib.Diff(split(before, stripTrailingCR), split(after, stripTrailingCR))
}

func split(value string, stripTrailingCR bool) []string {
	const sep = "\n"
	split := strings.Split(value, sep)
	if !stripTrailingCR {
		return split
	}
	var stripped []string
	for _, s := range split {
		stripped = append(stripped, strings.TrimSuffix(s, "\r"))
	}
	return stripped
}

func printDiffRecords(suppressedKinds []string, kind string, context int, diffs []difflib.DiffRecord, to io.Writer) {
	for _, ckind := range suppressedKinds {
		if ckind == kind {
			str := fmt.Sprintf("+ Changes suppressed on sensitive content of type %s\n", kind)
			_, _ = fmt.Fprint(to, ansi.Color(str, "yellow"))
			return
		}
	}

	if context >= 0 {
		distances := calculateDistances(diffs)
		omitting := false
		for i, diff := range diffs {
			if distances[i] > context {
				if !omitting {
					_, _ = fmt.Fprintln(to, "...")
					omitting = true
				}
			} else {
				omitting = false
				printDiffRecord(diff, to)
			}
		}
	} else {
		for _, diff := range diffs {
			printDiffRecord(diff, to)
		}
	}
}

func printDiffRecord(diff difflib.DiffRecord, to io.Writer) {
	text := diff.Payload

	switch diff.Delta {
	case difflib.RightOnly:
		_, _ = fmt.Fprintf(to, "%s\n", ansi.Color("+ "+text, "green"))
	case difflib.LeftOnly:
		_, _ = fmt.Fprintf(to, "%s\n", ansi.Color("- "+text, "red"))
	case difflib.Common:
		if text == "" {
			_, _ = fmt.Fprintln(to)
		} else {
			_, _ = fmt.Fprintf(to, "%s\n", "  "+text)
		}
	}
}

// calculateDistances returns the distance of every diff line to the closest change.
func calculateDistances(diffs []difflib.DiffRecord) map[int]int {
	distances := map[int]int{}

	change := -1
	for i, diff := range diffs {
		if diff.Delta != difflib.Common {
			change = i
		}
		distance := math.MaxInt32
		if change != -1 {
			distance = i - change
		}
		distances[i] = distance
	}

	change = -1
	for i := len(diffs) - 1; i >= 0; i-- {
		diff := diffs[i]
		if diff.Delta != difflib.Common {
			change = i
		}
		if change != -1 {
			distance := change - i
			if distance < distances[i] {
				distances[i] = distance
			}
		}
	}

	return distances
}

func sortedKeys(manifests map[string]*MappingResult) []string {
	var keys []string

	for key := range manifests {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
