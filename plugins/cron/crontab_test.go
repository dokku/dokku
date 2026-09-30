package cron

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	base36 "github.com/multiformats/go-base36"
)

func TestParseInjectedCronEntries(t *testing.T) {
	cases := []struct {
		name        string
		output      string
		wantTasks   []CronTask
		wantInvalid []string
	}{
		{
			name:   "text entry without a log file",
			output: "@daily;/bin/true",
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true", AltCommand: "/bin/true"},
			},
		},
		{
			name:   "text entry with a log file",
			output: "@daily;/bin/true;/var/log/dokku/log.log",
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true", AltCommand: "/bin/true", LogFile: "/var/log/dokku/log.log"},
			},
		},
		{
			name:   "json entry without a log file or mailto",
			output: `{"schedule":"@daily","command":"/bin/true"}`,
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true", AltCommand: "/bin/true"},
			},
		},
		{
			name:   "json entry with a log file and mailto",
			output: `{"schedule":"@daily","command":"/bin/true","log-file":"/var/log/dokku/log.log","mailto":"ops@example.com"}`,
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true", AltCommand: "/bin/true", LogFile: "/var/log/dokku/log.log", Mailto: "ops@example.com"},
			},
		},
		{
			name:   "json entry whose command holds a semicolon",
			output: `{"schedule":"@daily","command":"/bin/true; /bin/false"}`,
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true; /bin/false", AltCommand: "/bin/true; /bin/false"},
			},
		},
		{
			name:   "json and text entries printed together, with blank lines",
			output: "@daily;/bin/true\n\n" + `{"schedule":"@hourly","command":"/bin/false","mailto":"ops@example.com"}` + "\n",
			wantTasks: []CronTask{
				{Schedule: "@daily", Command: "/bin/true", AltCommand: "/bin/true"},
				{Schedule: "@hourly", Command: "/bin/false", AltCommand: "/bin/false", Mailto: "ops@example.com"},
			},
		},
		{
			name:        "text entry with too many fields",
			output:      "@daily;/bin/true;/var/log/dokku/log.log;ops@example.com",
			wantTasks:   []CronTask{},
			wantInvalid: []string{"@daily;/bin/true;/var/log/dokku/log.log;ops@example.com"},
		},
		{
			name:        "text entry with too few fields",
			output:      "@daily",
			wantTasks:   []CronTask{},
			wantInvalid: []string{"@daily"},
		},
		{
			name:        "json entry that does not decode",
			output:      `{"schedule":"@daily",`,
			wantTasks:   []CronTask{},
			wantInvalid: []string{`{"schedule":"@daily",`},
		},
		{
			name:        "json entry without a schedule",
			output:      `{"command":"/bin/true"}`,
			wantTasks:   []CronTask{},
			wantInvalid: []string{`{"command":"/bin/true"}`},
		},
		{
			name:        "json entry without a command",
			output:      `{"schedule":"@daily"}`,
			wantTasks:   []CronTask{},
			wantInvalid: []string{`{"schedule":"@daily"}`},
		},
		{
			name:        "json entry whose mailto holds a newline",
			output:      `{"schedule":"@daily","command":"/bin/true","mailto":"ops@example.com\n* * * * * /bin/evil"}`,
			wantTasks:   []CronTask{},
			wantInvalid: []string{`{"schedule":"@daily","command":"/bin/true","mailto":"ops@example.com\n* * * * * /bin/evil"}`},
		},
		{
			name:        "json entry whose command holds a carriage return",
			output:      `{"schedule":"@daily","command":"/bin/true\r/bin/evil"}`,
			wantTasks:   []CronTask{},
			wantInvalid: []string{`{"schedule":"@daily","command":"/bin/true\r/bin/evil"}`},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tasks, invalid := parseInjectedCronEntries(tc.output)
			for i := range tasks {
				if tasks[i].ID == "" {
					t.Errorf("task %d has no id", i)
				}
				tasks[i].ID = ""
			}

			if !reflect.DeepEqual(tasks, tc.wantTasks) {
				t.Errorf("parseInjectedCronEntries() tasks = %#v, want %#v", tasks, tc.wantTasks)
			}

			wantInvalid := tc.wantInvalid
			if wantInvalid == nil {
				wantInvalid = []string{}
			}
			if !reflect.DeepEqual(invalid, wantInvalid) {
				t.Errorf("parseInjectedCronEntries() invalid = %#v, want %#v", invalid, wantInvalid)
			}
		})
	}
}

// TestParseInjectedCronEntriesKeepsTextIDs pins that the id of a text entry is
// generated as it was before json entries were read, and that a json entry
// without a mailto gets the same id as the text entry it replaces.
func TestParseInjectedCronEntriesKeepsTextIDs(t *testing.T) {
	want := base36.EncodeToStringLc([]byte("@daily;;;/bin/true;;;/var/log/dokku/log.log"))

	for _, output := range []string{
		"@daily;/bin/true;/var/log/dokku/log.log",
		`{"schedule":"@daily","command":"/bin/true","log-file":"/var/log/dokku/log.log"}`,
	} {
		tasks, _ := parseInjectedCronEntries(output)
		if len(tasks) != 1 {
			t.Fatalf("parseInjectedCronEntries(%q) returned %d tasks, want 1", output, len(tasks))
		}
		if tasks[0].ID != want {
			t.Errorf("parseInjectedCronEntries(%q) id = %q, want %q", output, tasks[0].ID, want)
		}
	}

	withMailto, _ := parseInjectedCronEntries(`{"schedule":"@daily","command":"/bin/true","log-file":"/var/log/dokku/log.log","mailto":"ops@example.com"}`)
	if len(withMailto) != 1 || withMailto[0].ID == want {
		t.Errorf("a json entry with a mailto should not share the id of the entry without one")
	}
}

// renderCronTemplate renders the cron template for the given tasks and global
// MAILTO
func renderCronTemplate(t *testing.T, tasks []CronTask, mailto string) string {
	t.Helper()

	tmpl, err := getCronTemplate()
	if err != nil {
		t.Fatalf("getCronTemplate() returned an error: %v", err)
	}

	var got bytes.Buffer
	if err := tmpl.Execute(&got, cronTemplateData(tasks, "", mailto)); err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}

	return got.String()
}

// TestCronTemplateResetsMailtoAfterATaskWithItsOwn pins that a task with its own
// MAILTO is written in place between a MAILTO line of its own and one resetting
// it to the global MAILTO, so no task after it is mailed to its recipient.
func TestCronTemplateResetsMailtoAfterATaskWithItsOwn(t *testing.T) {
	tasks := []CronTask{
		{Schedule: "@daily", AltCommand: "/bin/true", LogFile: "/var/log/dokku/log.log"},
		{Schedule: "@hourly", AltCommand: "/bin/alpha", LogFile: "/var/log/dokku/alpha.log", Mailto: "alpha@example.com"},
		{Schedule: "@daily", AltCommand: "/bin/false"},
		{Schedule: "@daily", AltCommand: "/bin/zeta", LogFile: "/var/log/dokku/zeta.log", Mailto: "zeta@example.com"},
		{Schedule: "@weekly", AltCommand: "/bin/zeta-weekly", Mailto: "zeta@example.com"},
	}

	want := strings.Join([]string{
		"MAILTO=global@example.com",
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"SHELL=/bin/bash",
		"",
		"@daily /bin/true &>> /var/log/dokku/log.log",
		"MAILTO=alpha@example.com",
		"@hourly /bin/alpha 2>&1 | tee -a /var/log/dokku/alpha.log",
		"MAILTO=global@example.com",
		"@daily /bin/false",
		"MAILTO=zeta@example.com",
		"@daily /bin/zeta 2>&1 | tee -a /var/log/dokku/zeta.log",
		"MAILTO=global@example.com",
		"MAILTO=zeta@example.com",
		"@weekly /bin/zeta-weekly",
		"MAILTO=global@example.com",
		"",
	}, "\n")
	if got := renderCronTemplate(t, tasks, "global@example.com"); got != want {
		t.Errorf("rendered crontab =\n%s\nwant\n%s", got, want)
	}
}

// TestCronTemplateResetsMailtoToTheCrontabOwner pins that without a global
// MAILTO, a task with its own is followed by a MAILTO naming the crontab's
// owner, who cron mails when MAILTO is unset. An empty MAILTO would stop cron
// mailing anyone about the tasks after it.
func TestCronTemplateResetsMailtoToTheCrontabOwner(t *testing.T) {
	tasks := []CronTask{
		{Schedule: "@hourly", AltCommand: "/bin/alpha", LogFile: "/var/log/dokku/alpha.log", Mailto: "alpha@example.com"},
		{Schedule: "@daily", AltCommand: "/bin/true", LogFile: "/var/log/dokku/log.log"},
	}

	want := strings.Join([]string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"SHELL=/bin/bash",
		"",
		"MAILTO=alpha@example.com",
		"@hourly /bin/alpha 2>&1 | tee -a /var/log/dokku/alpha.log",
		"MAILTO=dokku",
		"@daily /bin/true &>> /var/log/dokku/log.log",
		"",
	}, "\n")
	if got := renderCronTemplate(t, tasks, ""); got != want {
		t.Errorf("rendered crontab =\n%s\nwant\n%s", got, want)
	}
}

// TestCronTemplateWithoutTaskMailto pins that a crontab without tasks that set
// their own MAILTO renders as it did before per-task MAILTO values existed.
func TestCronTemplateWithoutTaskMailto(t *testing.T) {
	tasks := []CronTask{
		{Schedule: "@daily", AltCommand: "/bin/true", LogFile: "/var/log/dokku/log.log"},
	}

	got := renderCronTemplate(t, tasks, "")

	want := strings.Join([]string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"SHELL=/bin/bash",
		"",
		"@daily /bin/true &>> /var/log/dokku/log.log",
		"",
	}, "\n")
	if got != want {
		t.Errorf("rendered crontab =\n%s\nwant\n%s", got, want)
	}
}
