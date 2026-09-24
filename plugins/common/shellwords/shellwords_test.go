package shellwords

import (
	"slices"
	"testing"
)

func TestSplit(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "empty input",
			input: "",
		},
		{
			name:  "plain words",
			input: "python -m http.server 5000",
			want:  []string{"python", "-m", "http.server", "5000"},
		},
		{
			name:  "double quoted word",
			input: `sh -c "echo hi"`,
			want:  []string{"sh", "-c", "echo hi"},
		},
		{
			name:  "single quoted word",
			input: `sh -c 'echo hi'`,
			want:  []string{"sh", "-c", "echo hi"},
		},
		{
			name:  "escaped quotes inside double quotes",
			input: `sh -c "echo \"hi\""`,
			want:  []string{"sh", "-c", `echo "hi"`},
		},
		{
			name:  "escaped single quote outside quotes",
			input: `echo it\'s`,
			want:  []string{"echo", "it's"},
		},
		{
			name:  "parameter expansion kept verbatim",
			input: `sh -c "echo $PORT"`,
			want:  []string{"sh", "-c", "echo $PORT"},
		},
		{
			name:  "command substitution kept verbatim",
			input: `echo $(date)`,
			want:  []string{"echo", "$(date)"},
		},
		{
			name:  "backticks kept verbatim",
			input: "echo `date`",
			want:  []string{"echo", "`date`"},
		},
		{
			name:  "escaped backticks inside double quotes",
			input: "echo \"\\`date\\`\"",
			want:  []string{"echo", "`date`"},
		},
		{
			name:    "unterminated quote",
			input:   `sh -c "echo hi`,
			wantErr: true,
		},
		{
			name:    "bare operator",
			input:   `a && b`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Split(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Split(%q) expected error, got %q", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Split(%q) unexpected error: %v", tc.input, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("Split(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSplitWithEndReportsSkippedComment(t *testing.T) {
	input := "a # comment"
	fields, end, err := SplitWithEnd(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Equal(fields, []string{"a"}) {
		t.Fatalf("fields = %q, want [a]", fields)
	}
	if end != 1 {
		t.Fatalf("end = %d, want 1", end)
	}
}
