package dockeroptions

import (
	"testing"
)

func TestSplitOptionString(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		wantOptions   []string
		wantProcesses []string
		wantErr       bool
	}{
		{
			name:        "single flag with value",
			input:       "--link foo",
			wantOptions: []string{"--link foo"},
		},
		{
			name:        "single flag with equals value",
			input:       "--build-arg FOO=bar",
			wantOptions: []string{"--build-arg FOO=bar"},
		},
		{
			name:        "two link flags split",
			input:       "--link a --link b",
			wantOptions: []string{"--link a", "--link b"},
		},
		{
			name:        "build-arg followed by two link flags",
			input:       "--build-arg X=Y --link a --link b",
			wantOptions: []string{"--build-arg X=Y", "--link a", "--link b"},
		},
		{
			name:        "boolean flag then key+value flag",
			input:       "--rm --link foo",
			wantOptions: []string{"--rm", "--link foo"},
		},
		{
			name:        "short flag with value",
			input:       "-v /tmp",
			wantOptions: []string{"-v /tmp"},
		},
		{
			name:        "restart with embedded equals and colon",
			input:       "--restart=on-failure:5",
			wantOptions: []string{"--restart=on-failure:5"},
		},
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   \t  ",
		},
		{
			name:        "value with whitespace gets shell-quoted",
			input:       `--label "foo bar"`,
			wantOptions: []string{"--label 'foo bar'"},
		},
		{
			name:        "equals value with whitespace gets shell-quoted",
			input:       `--label key="hello world"`,
			wantOptions: []string{"--label 'key=hello world'"},
		},
		{
			name:          "process flag in option content is lifted",
			input:         "--process web",
			wantProcesses: []string{"web"},
		},
		{
			name:          "process equals form is lifted",
			input:         "--process=web",
			wantProcesses: []string{"web"},
		},
		{
			name:          "process lifted from middle of multi-flag input",
			input:         "--link foo --process web --link bar",
			wantOptions:   []string{"--link foo", "--link bar"},
			wantProcesses: []string{"web"},
		},
		{
			name:          "multiple processes lifted",
			input:         "--process web --process worker",
			wantProcesses: []string{"web", "worker"},
		},
		{
			name:    "process without value errors",
			input:   "--process",
			wantErr: true,
		},
		{
			name:    "unbalanced quote errors",
			input:   `--build-arg FOO="bar`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotOptions, gotProcesses, err := SplitOptionString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("SplitOptionString(%q) succeeded, want error", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("SplitOptionString(%q): %v", tc.input, err)
			}
			if !equalStringSlice(gotOptions, tc.wantOptions) {
				t.Errorf("options = %q, want %q", gotOptions, tc.wantOptions)
			}
			if !equalStringSlice(gotProcesses, tc.wantProcesses) {
				t.Errorf("processes = %q, want %q", gotProcesses, tc.wantProcesses)
			}
		})
	}
}

func TestQuoteShellArg(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "''"},
		{"plain", "plain"},
		{"FOO=bar", "FOO=bar"},
		{"hello world", "'hello world'"},
		{"it's", `'it'\''s'`},
		{"$VAR", "'$VAR'"},
		{"a|b", "'a|b'"},
	}
	for _, tc := range cases {
		if got := quoteShellArg(tc.in); got != tc.want {
			t.Errorf("quoteShellArg(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
