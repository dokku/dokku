package common

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestReportSingleAppInputValidate(t *testing.T) {
	t.Parallel()

	baseValid := ReportSingleAppInput{
		ReportType: "ps",
		AppName:    "myapp",
		InfoFlags:  map[string]string{},
		Format:     "stdout",
	}

	tests := []struct {
		name      string
		input     ReportSingleAppInput
		wantError string
	}{
		{
			name: "valid",
			input: ReportSingleAppInput{
				ReportType: "ps",
				AppName:    "myapp",
				InfoFlags:  map[string]string{"--ps-restart-policy": "always"},
				Format:     "stdout",
			},
		},
		{
			name: "valid empty format defaults to stdout",
			input: ReportSingleAppInput{
				ReportType: "ps",
				AppName:    "myapp",
				InfoFlags:  map[string]string{},
				Format:     "",
			},
		},
		{
			name:      "missing ReportType",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.ReportType = "" }),
			wantError: "ReportType is required",
		},
		{
			name:      "ReportType with space rejected",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.ReportType = "docker options" }),
			wantError: "ReportType must not contain whitespace",
		},
		{
			name:      "ReportType with tab rejected",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.ReportType = "foo\tbar" }),
			wantError: "ReportType must not contain whitespace",
		},
		{
			name:      "ReportType beginning with -- rejected",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.ReportType = "--ps" }),
			wantError: "ReportType must not begin with --",
		},
		{
			name:      "missing AppName",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.AppName = "" }),
			wantError: "AppName is required",
		},
		{
			name:      "nil InfoFlags",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.InfoFlags = nil }),
			wantError: "InfoFlags is required",
		},
		{
			name:      "invalid Format",
			input:     mut(baseValid, func(i *ReportSingleAppInput) { i.Format = "yaml" }),
			wantError: "Format must be",
		},
		{
			name: "json with info flag rejected",
			input: ReportSingleAppInput{
				ReportType: "ps",
				AppName:    "myapp",
				InfoFlags:  map[string]string{},
				Format:     "json",
				InfoFlag:   "--ps-restart-policy",
			},
			wantError: "--format flag cannot be specified when specifying an info flag",
		},
		{
			name: "EmitLegacyPrefix without TrimPrefix rejected",
			input: ReportSingleAppInput{
				ReportType:       "ps",
				AppName:          "myapp",
				InfoFlags:        map[string]string{},
				Format:           "stdout",
				EmitLegacyPrefix: true,
				TrimPrefix:       false,
			},
			wantError: "EmitLegacyPrefix has no effect when TrimPrefix is false",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.input.Validate()
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantError)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("expected error to contain %q, got %q", tc.wantError, err.Error())
			}
		})
	}
}

func TestReportSingleAppJSONKeys(t *testing.T) {
	infoFlags := map[string]string{
		"--ps-stop-timeout-seconds": "30",
		"--ps-restart-policy":       "always",
		"--deployed":                "true",
	}

	tests := []struct {
		name             string
		trimPrefix       bool
		emitLegacyPrefix bool
		wantKeys         map[string]string
	}{
		{
			name:             "trimPrefix=false: prefixed keys only",
			trimPrefix:       false,
			emitLegacyPrefix: false,
			wantKeys: map[string]string{
				"ps-stop-timeout-seconds": "30",
				"ps-restart-policy":       "always",
				"deployed":                "true",
			},
		},
		{
			name:             "trimPrefix=true emitLegacy=false: stripped keys only",
			trimPrefix:       true,
			emitLegacyPrefix: false,
			wantKeys: map[string]string{
				"stop-timeout-seconds": "30",
				"restart-policy":       "always",
				"deployed":             "true",
			},
		},
		{
			name:             "trimPrefix=true emitLegacy=true: both shapes",
			trimPrefix:       true,
			emitLegacyPrefix: true,
			wantKeys: map[string]string{
				"stop-timeout-seconds":    "30",
				"restart-policy":          "always",
				"deployed":                "true",
				"ps-stop-timeout-seconds": "30",
				"ps-restart-policy":       "always",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				err := ReportSingleApp(ReportSingleAppInput{
					ReportType:       "ps",
					AppName:          "myapp",
					InfoFlags:        infoFlags,
					Format:           "json",
					TrimPrefix:       tc.trimPrefix,
					EmitLegacyPrefix: tc.emitLegacyPrefix,
				})
				if err != nil {
					t.Fatalf("ReportSingleApp returned error: %v", err)
				}
			})

			var got map[string]string
			if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
			}

			if len(got) != len(tc.wantKeys) {
				t.Fatalf("expected %d keys, got %d\nwant: %v\ngot:  %v", len(tc.wantKeys), len(got), tc.wantKeys, got)
			}
			for k, v := range tc.wantKeys {
				if got[k] != v {
					t.Errorf("expected %q=%q, got %q=%q", k, v, k, got[k])
				}
			}
		})
	}
}

func TestReportSingleAppJSONNonPluginPrefixKey(t *testing.T) {
	// Keys that don't start with --<reportType>- should pass through with
	// only the leading "--" stripped, regardless of TrimPrefix.
	infoFlags := map[string]string{
		"--build-id":          "abc123",
		"--builds-retention":  "5",
	}

	out := captureStdout(t, func() {
		err := ReportSingleApp(ReportSingleAppInput{
			ReportType:       "builds",
			AppName:          "myapp",
			InfoFlags:        infoFlags,
			Format:           "json",
			TrimPrefix:       true,
			EmitLegacyPrefix: true,
		})
		if err != nil {
			t.Fatalf("ReportSingleApp returned error: %v", err)
		}
	})

	var got map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if got["build-id"] != "abc123" {
		t.Errorf("expected build-id=abc123, got %q", got["build-id"])
	}
	if got["retention"] != "5" {
		t.Errorf("expected retention=5 (stripped), got %q", got["retention"])
	}
	if got["builds-retention"] != "5" {
		t.Errorf("expected builds-retention=5 (legacy), got %q", got["builds-retention"])
	}
	if _, present := got["uilds-retention"]; present {
		t.Error("found malformed key uilds-retention - prefix stripping was too aggressive")
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

func mut(base ReportSingleAppInput, f func(*ReportSingleAppInput)) ReportSingleAppInput {
	cp := base
	f(&cp)
	return cp
}
