package cron

import (
	"strings"
	"testing"

	appjson "github.com/dokku/dokku/plugins/app-json"
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

func TestDokkuRunCommandStdoutLogFile(t *testing.T) {
	task := CronTask{
		App:           "myapp",
		ID:            "abc123",
		StdoutLogFile: "/var/log/dokku/out.log",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123 >> /var/log/dokku/out.log"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestDokkuRunCommandStderrLogFile(t *testing.T) {
	task := CronTask{
		App:           "myapp",
		ID:            "abc123",
		StderrLogFile: "/var/log/dokku/err.log",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123 2>> /var/log/dokku/err.log"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestDokkuRunCommandSeparateStdoutAndStderrLogFiles(t *testing.T) {
	task := CronTask{
		App:           "myapp",
		ID:            "abc123",
		StdoutLogFile: "/var/log/dokku/out.log",
		StderrLogFile: "/var/log/dokku/err.log",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123 >> /var/log/dokku/out.log 2>> /var/log/dokku/err.log"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestDokkuRunCommandMergedLogFile(t *testing.T) {
	task := CronTask{
		App:     "myapp",
		ID:      "abc123",
		LogFile: "/var/log/dokku/merged.log",
	}

	got := task.DokkuRunCommand()
	want := "dokku cron:run myapp abc123 &>> /var/log/dokku/merged.log"
	if got != want {
		t.Errorf("DokkuRunCommand() = %q, want %q", got, want)
	}
}

func TestValidateCronLogPathAcceptsSafePaths(t *testing.T) {
	cases := []string{
		"/var/log/dokku/app.log",
		"/var/log/dokku/app-task_1.log",
		"/tmp/out.log",
	}
	for _, path := range cases {
		if err := ValidateCronLogPath(path); err != nil {
			t.Errorf("ValidateCronLogPath(%q) returned error: %v", path, err)
		}
	}
}

func TestValidateCronLogPathRejectsUnsafePaths(t *testing.T) {
	cases := []string{
		"",                         // empty
		"relative/path.log",        // not absolute
		"/var/log/dokku/a b.log",   // whitespace
		"/var/log/$(whoami).log",   // command substitution
		"/var/log/x.log; rm -rf /", // command injection
		"/var/log/x.log | cat",     // pipe
		"/var/log/x&.log",          // ampersand
		"/var/log/`id`.log",        // backticks
		"/var/log/*.log",           // glob
		"/var/log/x.log\nnewline",  // newline
		"/var/log/\"quoted\".log",  // quote
	}
	for _, path := range cases {
		if err := ValidateCronLogPath(path); err == nil {
			t.Errorf("ValidateCronLogPath(%q) accepted an unsafe path", path)
		}
	}
}

func fetchTasksForCron(t *testing.T, c appjson.CronTask, warnToFailure bool) ([]CronTask, error) {
	t.Helper()
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	return FetchCronTasks(FetchCronTasksInput{
		AppName:       "testapp",
		WarnToFailure: warnToFailure,
		AppJSON: &appjson.AppJSON{
			Cron: []appjson.CronTask{c},
		},
	})
}

func TestFetchCronTasksWiresSeparateLogFiles(t *testing.T) {
	tasks, err := fetchTasksForCron(t, appjson.CronTask{
		Command:       "npm run send-email",
		Schedule:      "@daily",
		StdoutLogFile: "/var/log/dokku/out.log",
		StderrLogFile: "/var/log/dokku/err.log",
	}, true)
	if err != nil {
		t.Fatalf("FetchCronTasks returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].StdoutLogFile != "/var/log/dokku/out.log" {
		t.Errorf("StdoutLogFile = %q, want /var/log/dokku/out.log", tasks[0].StdoutLogFile)
	}
	if tasks[0].StderrLogFile != "/var/log/dokku/err.log" {
		t.Errorf("StderrLogFile = %q, want /var/log/dokku/err.log", tasks[0].StderrLogFile)
	}
}

func TestFetchCronTasksWiresMergedLogFile(t *testing.T) {
	tasks, err := fetchTasksForCron(t, appjson.CronTask{
		Command:  "npm run send-email",
		Schedule: "@daily",
		LogFile:  "/var/log/dokku/merged.log",
	}, true)
	if err != nil {
		t.Fatalf("FetchCronTasks returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].LogFile != "/var/log/dokku/merged.log" {
		t.Errorf("LogFile = %q, want /var/log/dokku/merged.log", tasks[0].LogFile)
	}
}

func TestFetchCronTasksRejectsLogFileWithSeparateLogFiles(t *testing.T) {
	_, err := fetchTasksForCron(t, appjson.CronTask{
		Command:       "npm run send-email",
		Schedule:      "@daily",
		LogFile:       "/var/log/dokku/merged.log",
		StdoutLogFile: "/var/log/dokku/out.log",
	}, true)
	if err == nil {
		t.Fatalf("expected error for mutually exclusive log files, got nil")
	}
	if !strings.Contains(err.Error(), "logfile") {
		t.Errorf("error = %q, want it to mention the mutual exclusion", err.Error())
	}
}

func TestFetchCronTasksMutualExclusionSkipsWhenNotFailing(t *testing.T) {
	tasks, err := fetchTasksForCron(t, appjson.CronTask{
		Command:       "npm run send-email",
		Schedule:      "@daily",
		LogFile:       "/var/log/dokku/merged.log",
		StderrLogFile: "/var/log/dokku/err.log",
	}, false)
	if err != nil {
		t.Fatalf("FetchCronTasks returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected the conflicting task to be skipped, got %d tasks", len(tasks))
	}
}

func TestFetchCronTasksRejectsUnsafeLogPath(t *testing.T) {
	_, err := fetchTasksForCron(t, appjson.CronTask{
		Command:       "npm run send-email",
		Schedule:      "@daily",
		StdoutLogFile: "/var/log/$(whoami).log",
	}, true)
	if err == nil {
		t.Fatalf("expected error for unsafe log path, got nil")
	}
	if !strings.Contains(err.Error(), "stdout_logfile") {
		t.Errorf("error = %q, want it to identify the offending field", err.Error())
	}
}

func TestFetchCronTasksUnsafeLogPathSkipsWhenNotFailing(t *testing.T) {
	tasks, err := fetchTasksForCron(t, appjson.CronTask{
		Command:       "npm run send-email",
		Schedule:      "@daily",
		StderrLogFile: "/var/log/x.log; rm -rf /",
	}, false)
	if err != nil {
		t.Fatalf("FetchCronTasks returned error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected the invalid task to be skipped, got %d tasks", len(tasks))
	}
}
