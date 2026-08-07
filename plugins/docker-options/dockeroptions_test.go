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
		{
			name:        "command substitution is stored verbatim",
			input:       "--label x=$(id)",
			wantOptions: []string{"--label 'x=$(id)'"},
		},
		{
			name:        "backtick command substitution is stored verbatim",
			input:       "--label x=`id`",
			wantOptions: []string{"--label 'x=`id`'"},
		},
		{
			name:        "parameter expansion is stored verbatim",
			input:       "--label x=$FOO",
			wantOptions: []string{"--label 'x=$FOO'"},
		},
		{
			name:        "metacharacter value that parses is shell-quoted",
			input:       `--label 'a;b'`,
			wantOptions: []string{"--label 'a;b'"},
		},
		{
			name:        "escaped backtick inside double quotes is unescaped",
			input:       "--label \"traefik.rule=Host(\\`app.example.com\\`)\"",
			wantOptions: []string{"--label 'traefik.rule=Host(`app.example.com`)'"},
		},
		{
			name:        "escaped dollar inside double quotes is unescaped",
			input:       "--label \"k=a\\$b\"",
			wantOptions: []string{"--label 'k=a$b'"},
		},
		{
			name:        "unquoted backslash escape is removed",
			input:       "--label k=a\\ b",
			wantOptions: []string{"--label 'k=a b'"},
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

func TestOptionsEqual(t *testing.T) {
	cases := []struct {
		name   string
		stored string
		option string
		want   bool
	}{
		{
			name:   "identical strings",
			stored: "-v /tmp:/tmp",
			option: "-v /tmp:/tmp",
			want:   true,
		},
		{
			name:   "legacy unquoted command substitution matches canonical form",
			stored: "--group-add $(getent group docker | cut -d: -f3)",
			option: "--group-add '$(getent group docker | cut -d: -f3)'",
			want:   true,
		},
		{
			name:   "legacy unquoted parameter expansion matches canonical form",
			stored: "--label FOO=$BAR",
			option: "--label 'FOO=$BAR'",
			want:   true,
		},
		{
			name:   "double quoted matches single quoted",
			stored: `--label "hello world"`,
			option: "--label 'hello world'",
			want:   true,
		},
		{
			name:   "different values do not match",
			stored: "-v /tmp:/tmp",
			option: "-v /var:/var",
			want:   false,
		},
		{
			name:   "different flags do not match",
			stored: "--link foo",
			option: "--link bar",
			want:   false,
		},
		{
			name:   "multi-flag entry does not match a single flag",
			stored: "-v /a:/a -v /b:/b",
			option: "-v /a:/a",
			want:   false,
		},
		{
			name:   "unbalanced quote never matches",
			stored: "--label 'unbalanced",
			option: "--label unbalanced",
			want:   false,
		},
		{
			name:   "empty stored entry does not match an option",
			stored: "",
			option: "-v /tmp:/tmp",
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := optionsEqual(tc.stored, tc.option); got != tc.want {
				t.Errorf("optionsEqual(%q, %q) = %v, want %v", tc.stored, tc.option, got, tc.want)
			}
		})
	}
}

func TestCanonicalOptionsFromLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "quotes an unquoted command substitution",
			line: "--group-add $(getent group docker | cut -d: -f3)",
			want: []string{"--group-add '$(getent group docker | cut -d: -f3)'"},
		},
		{
			name: "splits a line carrying several flags",
			line: "-v /a:/a -v /b:/b",
			want: []string{"-v /a:/a", "-v /b:/b"},
		},
		{
			name: "leaves an already canonical line untouched",
			line: "--group-add '$(getent group docker | cut -d: -f3)'",
			want: []string{"--group-add '$(getent group docker | cut -d: -f3)'"},
		},
		{
			name: "leaves a plain option untouched",
			line: "--restart=on-failure:5",
			want: []string{"--restart=on-failure:5"},
		},
		{
			name: "preserves a line with an unbalanced quote",
			line: "--label 'unbalanced",
			want: []string{"--label 'unbalanced"},
		},
		{
			name: "preserves a line whose tail reads as a shell comment",
			line: "-v /a:/a # a note",
			want: []string{"-v /a:/a # a note"},
		},
		{
			name: "returns nothing for an empty line",
			line: "",
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := canonicalOptionsFromLine(tc.line)
			if !equalStringSlice(got, tc.want) {
				t.Errorf("canonicalOptionsFromLine(%q) = %q, want %q", tc.line, got, tc.want)
			}

			if len(tc.want) == 0 {
				return
			}

			// Canonicalization must be idempotent so the one-time repair
			// pass is a no-op on stores it has already rewritten.
			again := canonicalizeOptionLines(got)
			if !equalStringSlice(again, got) {
				t.Errorf("canonicalizing %q again = %q, want %q", got, again, got)
			}
		})
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
