package logs

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestSinkValueToConfigInputs(t *testing.T) {
	sink, err := SinkValueToConfig(SinkValueToConfigInput{
		SinkValue: "console://?encoding[codec]=json",
		Inputs:    []string{"docker-source:myapp"},
	})
	if err != nil {
		t.Fatalf("SinkValueToConfig() error = %v", err)
	}

	if sink["type"] != "console" {
		t.Errorf("type = %v, want console", sink["type"])
	}

	want := []string{"docker-source:myapp"}
	if !reflect.DeepEqual(sink["inputs"], want) {
		t.Errorf("inputs = %v, want %v", sink["inputs"], want)
	}
}

func TestSinkValueToConfigOmitsEmptyInputs(t *testing.T) {
	sink, err := SinkValueToConfig(SinkValueToConfigInput{
		SinkValue: "console://?encoding[codec]=json",
	})
	if err != nil {
		t.Fatalf("SinkValueToConfig() error = %v", err)
	}

	if _, ok := sink["inputs"]; ok {
		t.Errorf("inputs should be omitted when no inputs are supplied, got %v", sink["inputs"])
	}
}

func TestSinkValueToConfigRejectsSinksOption(t *testing.T) {
	_, err := SinkValueToConfig(SinkValueToConfigInput{
		SinkValue: "console://?sinks=nope",
	})
	if err == nil {
		t.Fatal("SinkValueToConfig() expected an error for the sinks option")
	}
}

func TestSinkValueToConfigSchemeUnderscores(t *testing.T) {
	for _, scheme := range []string{"datadog_logs", "aws_cloudwatch_logs"} {
		sink, err := SinkValueToConfig(SinkValueToConfigInput{
			SinkValue: scheme + "://?api_key=abc123",
		})
		if err != nil {
			t.Fatalf("SinkValueToConfig(%s) error = %v", scheme, err)
		}

		if sink["type"] != scheme {
			t.Errorf("type = %v, want %s", sink["type"], scheme)
		}
	}
}

func TestSinkValueToConfigBase64Enc(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("{{ pod }}"))
	sink, err := SinkValueToConfig(SinkValueToConfigInput{
		SinkValue: "http://?process=base64enc%3A" + encoded,
	})
	if err != nil {
		t.Fatalf("SinkValueToConfig() error = %v", err)
	}

	if sink["process"] != "{{ pod }}" {
		t.Errorf("process = %v, want {{ pod }}", sink["process"])
	}
}

// TestSinkValueToConfigTemplatedFilePath guards the DSN handling that the
// documented cron file sink depends on: a vector template in a query-string
// value must survive url.Parse and the qson decode with its braces and
// interior spaces intact.
func TestSinkValueToConfigTemplatedFilePath(t *testing.T) {
	sink, err := SinkValueToConfig(SinkValueToConfigInput{
		SinkValue: "file://?path=/var/log/dokku/apps/myapp/cron-{{ dokku_cron_id }}.log&encoding[codec]=text",
		Inputs:    []string{"docker-cron-remap:myapp"},
	})
	if err != nil {
		t.Fatalf("SinkValueToConfig() error = %v", err)
	}

	if sink["type"] != "file" {
		t.Errorf("type = %v, want file", sink["type"])
	}

	wantPath := "/var/log/dokku/apps/myapp/cron-{{ dokku_cron_id }}.log"
	if sink["path"] != wantPath {
		t.Errorf("path = %v, want %v", sink["path"], wantPath)
	}

	encoding, ok := sink["encoding"].(map[string]interface{})
	if !ok {
		t.Fatalf("encoding = %v, want a map", sink["encoding"])
	}
	if encoding["codec"] != "text" {
		t.Errorf("encoding.codec = %v, want text", encoding["codec"])
	}
}
