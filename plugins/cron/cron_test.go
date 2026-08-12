package cron

import (
	"strings"
	"testing"
)

func TestDokkuRunCommandAppTaskDispatchesViaCronRun(t *testing.T) {
	task := CronTask{
		App:               "myapp",
		ID:                "abc123",
		Command:           "echo CRON_OK; echo hi > /tmp/appjson-test.txt",
		Schedule:          "* * * * *",
		ConcurrencyPolicy: "allow",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}

	if strings.Contains(got, task.Command) {
		t.Errorf("DokkuRunCommand() leaked user command into crontab line: %q", got)
	}
	if strings.ContainsAny(got, ";>|&`$") {
		t.Errorf("DokkuRunCommand() contains shell metacharacters: %q", got)
	}
}

func TestDokkuRunCommandPlainCommandStillUsesCronRun(t *testing.T) {
	task := CronTask{
		App:               "myapp",
		ID:                "abc123",
		Command:           "npm run send-email",
		Schedule:          "@daily",
		ConcurrencyPolicy: "forbid",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestDokkuRunCommandAltCommandUnchanged(t *testing.T) {
	task := CronTask{
		ID:         "abc123",
		AltCommand: "/usr/bin/some-internal-task --flag",
	}

	got := task.DokkuRunCommand()
	want := "/usr/bin/some-internal-task --flag"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestValidateCronCommandAcceptsValidCommands(t *testing.T) {
	cases := []string{
		"python3 task.py schedule",
		"npm run send-email",
		"sh -c 'echo CRON_OK; echo hi > /tmp/x.txt'",
		`node -e 'console.log(1)'`,
		"true",
	}
	for _, cmd := range cases {
		if err := ValidateCronCommand(cmd); err != nil {
			t.Errorf("ValidateCronCommand(%q) returned error: %v", cmd, err)
		}
	}
}

func TestValidateCronCommandRejectsShellOperators(t *testing.T) {
	cases := []string{
		"echo CRON_OK; echo hi > /tmp/x.txt",
		"cmd1 && cmd2",
		"cmd | other",
		"cmd > file",
		"cmd $(other)",
	}
	for _, cmd := range cases {
		if err := ValidateCronCommand(cmd); err == nil {
			t.Errorf("ValidateCronCommand(%q) accepted a command containing a shell operator", cmd)
		}
	}
}

func TestDokkuRunCommandAltCommandWithLogFile(t *testing.T) {
	task := CronTask{
		ID:         "abc123",
		AltCommand: "/usr/bin/some-internal-task",
		LogFile:    "/var/log/dokku/internal-task.log",
	}

	got := task.DokkuRunCommand()
	want := "/usr/bin/some-internal-task &>> /var/log/dokku/internal-task.log"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

// TestDokkuRunCommandAppTaskIgnoresLogFile pins that LogFile is honored only
// for internally injected tasks from the cron-entries trigger. App tasks never
// interpolate a path into the crontab line - their output is shipped by the
// vector integration instead.
func TestDokkuRunCommandAppTaskIgnoresLogFile(t *testing.T) {
	task := CronTask{
		App:     "myapp",
		ID:      "abc123",
		Command: "npm run send-email",
		LogFile: "/var/log/dokku/should-not-appear.log",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}

	if strings.Contains(got, ">>") {
		t.Errorf("DokkuRunCommand() interpolated a redirect into an app task line: %q", got)
	}
}

// TestValidateTTLSeconds pins that a cron task deadline must be a positive
// number of seconds. A zero or negative value would either expire the task the
// instant it starts or leave it running forever, and the docker-local retire
// pass only reaps containers whose deadline has actually elapsed.
func TestValidateTTLSeconds(t *testing.T) {
	cases := []struct {
		name       string
		ttlSeconds int64
		wantErr    bool
	}{
		{name: "default", ttlSeconds: DefaultTTLSeconds},
		{name: "positive override", ttlSeconds: 1},
		{name: "zero", ttlSeconds: 0, wantErr: true},
		{name: "negative", ttlSeconds: -1, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTTLSeconds(tc.ttlSeconds)
			if tc.wantErr && err == nil {
				t.Errorf("validateTTLSeconds(%d) = nil, want an error", tc.ttlSeconds)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateTTLSeconds(%d) = %v, want nil", tc.ttlSeconds, err)
			}
		})
	}
}
