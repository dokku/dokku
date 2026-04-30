package schedulerdockerlocal

import (
	"strings"
	"testing"

	"github.com/dokku/dokku/plugins/storage"
)

func TestBuildDockerExecArgsTTY(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerDockerLocal,
		HostPath:  "/var/lib/dokku/data/storage/demo-data",
	}

	cases := []struct {
		name        string
		interactive bool
		tty         bool
		want        string
	}{
		{"non-interactive", false, false, ""},
		{"piped stdin", true, false, "-i"},
		{"interactive tty", true, true, "-it"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := buildDockerExecArgs(entry, StorageExecInput{
				Image:       "alpine:3",
				Interactive: tc.interactive,
				Tty:         tc.tty,
				Command:     []string{"ls", "/data"},
			})
			if err != nil {
				t.Fatalf("buildDockerExecArgs returned error: %v", err)
			}
			joined := strings.Join(args, " ")
			if tc.want == "" {
				if strings.Contains(joined, " -i ") || strings.Contains(joined, " -it ") {
					t.Fatalf("expected no -i / -it flag, got: %s", joined)
				}
			} else if !strings.Contains(joined, " "+tc.want+" ") {
				t.Fatalf("expected to find %q in args, got: %s", tc.want, joined)
			}
		})
	}
}

func TestBuildDockerExecArgsUser(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerDockerLocal,
		HostPath:  "/var/lib/dokku/data/storage/demo-data",
		Chown:     "false",
	}

	args, err := buildDockerExecArgs(entry, StorageExecInput{
		Image:   "alpine:3",
		Command: []string{"ls"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if strings.Contains(strings.Join(args, " "), "--user") {
		t.Fatalf("did not expect --user flag with chown=false, got: %v", args)
	}

	args, err = buildDockerExecArgs(entry, StorageExecInput{
		Image:   "alpine:3",
		AsUser:  "1234",
		Command: []string{"ls"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(strings.Join(args, " "), "--user 1234:1234") {
		t.Fatalf("expected --user 1234:1234, got: %v", args)
	}
}

func TestBuildDockerExecArgsDefaultsToShellWhenNoCommand(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerDockerLocal,
		HostPath:  "/var/lib/dokku/data/storage/demo-data",
	}
	args, err := buildDockerExecArgs(entry, StorageExecInput{
		Image: "alpine:3",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	last := args[len(args)-3:]
	if last[0] != "sh" || last[1] != "-c" {
		t.Fatalf("expected the trailing command to be a shell fallback, got: %v", last)
	}
}

func TestBuildDockerExecArgsMountsHostPath(t *testing.T) {
	entry := &storage.Entry{
		Name:      "demo-data",
		Scheduler: storage.SchedulerDockerLocal,
		HostPath:  "/srv/demo",
	}
	args, err := buildDockerExecArgs(entry, StorageExecInput{
		Image:   "alpine:3",
		Command: []string{"ls"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-v /srv/demo:/data") {
		t.Fatalf("expected -v /srv/demo:/data, got: %s", joined)
	}
}
