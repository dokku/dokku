package dockeroptions

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/common"
)

func setupReportEnv(t *testing.T, appName string) {
	t.Helper()
	dokkuRoot := setupMigrationEnv(t)
	if err := os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
}

func TestListReportKey(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"build", "build-list"},
		{"deploy", "deploy-list"},
		{"run", "run-list"},
		{"deploy.web", "deploy.web-list"},
	}
	for _, tc := range cases {
		if got := listReportKey(tc.in); got != tc.want {
			t.Errorf("listReportKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildJSONReportData_IncludesListKeysForShorthandOnly(t *testing.T) {
	const appName = "testapp"
	setupReportEnv(t, appName)

	if err := common.PropertyListWrite("docker-options", appName, "_default_.deploy", []string{"-v /logs:/logs", "--memory=512m"}); err != nil {
		t.Fatalf("PropertyListWrite deploy: %v", err)
	}
	if err := common.PropertyListWrite("docker-options", appName, "web.deploy", []string{"-p 8080:5000"}); err != nil {
		t.Fatalf("PropertyListWrite web deploy: %v", err)
	}

	infoFlags := map[string]string{
		"--docker-options-build":      "",
		"--docker-options-deploy":     "-v /logs:/logs --memory=512m",
		"--docker-options-run":        "",
		"--docker-options-deploy.web": "-p 8080:5000",
	}

	data, err := buildJSONReportData(appName, infoFlags)
	if err != nil {
		t.Fatalf("buildJSONReportData: %v", err)
	}

	assertStringValue(t, data, "deploy", "-v /logs:/logs --memory=512m")
	assertStringValue(t, data, "docker-options-deploy", "-v /logs:/logs --memory=512m")
	assertStringSliceValue(t, data, "deploy-list", []string{"-v /logs:/logs", "--memory=512m"})
	assertStringSliceValue(t, data, "build-list", []string{})
	assertStringSliceValue(t, data, "run-list", []string{})
	assertStringSliceValue(t, data, "deploy.web-list", []string{"-p 8080:5000"})

	if _, ok := data["docker-options-deploy-list"]; ok {
		t.Fatal("docker-options-deploy-list should not be present")
	}
}

func TestBuildJSONReportData_PreservesOptionsWithSpaces(t *testing.T) {
	const appName = "testapp"
	setupReportEnv(t, appName)

	option := `--label 'com.example.description=my app'`
	if err := common.PropertyListWrite("docker-options", appName, "_default_.deploy", []string{option, "--cpus=2"}); err != nil {
		t.Fatalf("PropertyListWrite deploy: %v", err)
	}

	infoFlags := map[string]string{
		"--docker-options-build":  "",
		"--docker-options-deploy": option + " --cpus=2",
		"--docker-options-run":    "",
	}

	data, err := buildJSONReportData(appName, infoFlags)
	if err != nil {
		t.Fatalf("buildJSONReportData: %v", err)
	}

	assertStringSliceValue(t, data, "deploy-list", []string{option, "--cpus=2"})
}

func TestBuildJSONReportData_GlobalReportHasNoListKeys(t *testing.T) {
	data, err := buildJSONReportData("--global", map[string]string{})
	if err != nil {
		t.Fatalf("buildJSONReportData: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("global report data = %#v, want empty map", data)
	}
}

func TestBuildJSONReportData_MarshalsEmptyListsAsArrays(t *testing.T) {
	const appName = "testapp"
	setupReportEnv(t, appName)

	infoFlags := map[string]string{
		"--docker-options-build":  "",
		"--docker-options-deploy": "",
		"--docker-options-run":    "",
	}

	data, err := buildJSONReportData(appName, infoFlags)
	if err != nil {
		t.Fatalf("buildJSONReportData: %v", err)
	}

	out, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	jsonText := string(out)
	for _, key := range []string{"build-list", "deploy-list", "run-list"} {
		if !strings.Contains(jsonText, `"`+key+`":[]`) {
			t.Errorf("json output missing empty array for %q: %s", key, jsonText)
		}
	}
}

func assertStringValue(t *testing.T, data map[string]any, key, want string) {
	t.Helper()
	got, ok := data[key]
	if !ok {
		t.Fatalf("key %q missing from report data", key)
	}
	gotString, ok := got.(string)
	if !ok {
		t.Fatalf("key %q = %#v, want string %q", key, got, want)
	}
	if gotString != want {
		t.Errorf("key %q = %q, want %q", key, gotString, want)
	}
}

func assertStringSliceValue(t *testing.T, data map[string]any, key string, want []string) {
	t.Helper()
	got, ok := data[key]
	if !ok {
		t.Fatalf("key %q missing from report data", key)
	}
	gotSlice, ok := got.([]string)
	if !ok {
		t.Fatalf("key %q = %#v, want []string %#v", key, got, want)
	}
	if !equalStringSlice(gotSlice, want) {
		t.Errorf("key %q = %#v, want %#v", key, gotSlice, want)
	}
}
