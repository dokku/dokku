package scheduler_k3s

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

func setupReportTest(t *testing.T, apps ...string) {
	t.Helper()
	t.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins")
	t.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	t.Setenv("DOKKU_SYSTEM_USER", "root")
	t.Setenv("DOKKU_SYSTEM_GROUP", "root")
	dokkuRoot := t.TempDir()
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	if err := common.PropertySetup("scheduler-k3s"); err != nil {
		t.Fatalf("PropertySetup: %v", err)
	}
	for _, appName := range apps {
		if err := os.MkdirAll(dokkuRoot+"/"+appName, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = orig
	return <-done
}

func TestExtractReportInfoFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prefix    string
		args      []string
		wantArgs  []string
		wantFlag  string
		wantError string
	}{
		{
			name:     "no info flag",
			prefix:   "--scheduler-k3s-annotations.",
			args:     []string{"myapp", "--format", "json"},
			wantArgs: []string{"myapp", "--format", "json"},
		},
		{
			name:     "single info flag",
			prefix:   "--scheduler-k3s-labels.",
			args:     []string{"myapp", "--scheduler-k3s-labels.global.deployment.foo"},
			wantArgs: []string{"myapp"},
			wantFlag: "--scheduler-k3s-labels.global.deployment.foo",
		},
		{
			name:      "multiple info flags",
			prefix:    "--scheduler-k3s-autoscaling-auth.",
			args:      []string{"myapp", "--scheduler-k3s-autoscaling-auth.datadog.apiKey", "--scheduler-k3s-autoscaling-auth.datadog.appKey"},
			wantError: "only a single info flag may be specified",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotArgs, gotFlag, err := ExtractReportInfoFlag(tc.prefix, tc.args)
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotFlag != tc.wantFlag {
				t.Fatalf("info flag = %q, want %q", gotFlag, tc.wantFlag)
			}
			if strings.Join(gotArgs, ",") != strings.Join(tc.wantArgs, ",") {
				t.Fatalf("passthrough args = %v, want %v", gotArgs, tc.wantArgs)
			}
		})
	}
}

func TestRenderProcessType(t *testing.T) {
	t.Parallel()

	if got := renderProcessType(GlobalProcessType); got != reportProcessTypeGlobal {
		t.Fatalf("renderProcessType(GlobalProcessType) = %q, want %q", got, reportProcessTypeGlobal)
	}
	if got := renderProcessType("web"); got != "web" {
		t.Fatalf("renderProcessType(\"web\") = %q, want %q", got, "web")
	}
}

func TestRenderAnnotationsLabelsReportJSON(t *testing.T) {
	out := captureStdout(t, func() {
		err := renderAnnotationsLabelsReport(annotationsLabelsRenderInput{
			AppName:    "myapp",
			ReportType: annotationsReportType,
			Heading:    "annotations",
			RowLabel:   "Annotation",
			Entries: []annotationsLabelsReportEntry{
				{ProcessType: GlobalProcessType, ResourceType: "deployment", Key: "foo", Value: "bar"},
				{ProcessType: "web", ResourceType: "deployment", Key: "baz", Value: "qux"},
			},
			Format: "json",
		})
		if err != nil {
			t.Fatalf("renderAnnotationsLabelsReport returned error: %v", err)
		}
	})

	var got map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}

	want := map[string]string{
		"global.deployment.foo": "bar",
		"web.deployment.baz":    "qux",
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("expected %q=%q, got %q=%q", key, value, key, got[key])
		}
	}
}

func TestRenderAutoscalingAuthReportJSON(t *testing.T) {
	out := captureStdout(t, func() {
		err := renderAutoscalingAuthReport(autoscalingAuthRenderInput{
			AppName: "myapp",
			Entries: []autoscalingAuthReportEntry{
				{Trigger: "datadog", MetadataKey: "apiKey", Value: "secret-1"},
				{Trigger: "datadog", MetadataKey: "appKey", Value: "secret-2"},
			},
			Format: "json",
		})
		if err != nil {
			t.Fatalf("renderAutoscalingAuthReport returned error: %v", err)
		}
	})

	var got map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}

	want := map[string]string{
		"datadog.apiKey": "secret-1",
		"datadog.appKey": "secret-2",
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("expected %q=%q, got %q=%q", key, value, key, got[key])
		}
	}
}

func TestReportAutoscalingAuthSingleAppInfoFlag(t *testing.T) {
	setupReportTest(t, "myapp")

	if err := common.PropertyWrite("scheduler-k3s", "myapp", "trigger-auth.datadog.apiKey", "secret-1"); err != nil {
		t.Fatalf("PropertyWrite: %v", err)
	}

	out := captureStdout(t, func() {
		err := ReportAutoscalingAuthSingleApp("myapp", "stdout", false, "--scheduler-k3s-autoscaling-auth.datadog.apiKey")
		if err != nil {
			t.Fatalf("ReportAutoscalingAuthSingleApp returned error: %v", err)
		}
	})

	if strings.TrimSpace(out) != tokenMask {
		t.Fatalf("info flag output = %q, want %q", strings.TrimSpace(out), tokenMask)
	}
}

func TestCollectAutoscalingAuthEntriesFromProperties(t *testing.T) {
	setupReportTest(t, "myapp")

	if err := common.PropertyWrite("scheduler-k3s", "myapp", "trigger-auth.datadog.apiKey", "secret-1"); err != nil {
		t.Fatalf("PropertyWrite apiKey: %v", err)
	}
	if err := common.PropertyWrite("scheduler-k3s", "myapp", "trigger-auth.datadog.appKey", "secret-2"); err != nil {
		t.Fatalf("PropertyWrite appKey: %v", err)
	}

	entries, err := collectAutoscalingAuthEntries("myapp")
	if err != nil {
		t.Fatalf("collectAutoscalingAuthEntries returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestReportAutoscalingAuthGlobalJSON(t *testing.T) {
	setupReportTest(t)

	if err := common.PropertyWrite("scheduler-k3s", "--global", "trigger-auth.datadog.apiKey", "global-secret"); err != nil {
		t.Fatalf("PropertyWrite: %v", err)
	}

	out := captureStdout(t, func() {
		err := ReportAutoscalingAuthSingleApp("--global", "json", false, "")
		if err != nil {
			t.Fatalf("ReportAutoscalingAuthSingleApp returned error: %v", err)
		}
	})

	var got map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}

	if got["datadog.apiKey"] != "global-secret" {
		t.Fatalf("expected datadog.apiKey=global-secret, got %v", got)
	}
}

func TestRenderAutoscalingAuthReportStdoutIncludesMetadataValues(t *testing.T) {
	out := captureStdout(t, func() {
		err := renderAutoscalingAuthReport(autoscalingAuthRenderInput{
			AppName: "myapp",
			Entries: []autoscalingAuthReportEntry{
				{Trigger: "datadog", MetadataKey: "apiKey", Value: "secret-1"},
			},
			Format:          "stdout",
			IncludeMetadata: true,
		})
		if err != nil {
			t.Fatalf("renderAutoscalingAuthReport returned error: %v", err)
		}
	})

	if !strings.Contains(out, "Datadog apiKey:") {
		t.Fatalf("expected metadata key in stdout, got: %s", out)
	}
	if !strings.Contains(out, "secret-1") {
		t.Fatalf("expected metadata value in stdout, got: %s", out)
	}
}
