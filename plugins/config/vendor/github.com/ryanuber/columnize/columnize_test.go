package columnize

import (
	"fmt"
	"reflect"
	"testing"

	crand "crypto/rand"
)

func TestListOfStringsInput(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyLinesOutput(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestLeadingSpacePreserved(t *testing.T) {
	input := []string{
		"| Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "   Column B  Column C\n"
	expected += "x  y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestColumnWidthCalculator(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"Longer than A | Longer than B | Longer than C",
		"short | short | short",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A       Column B       Column C\n"
	expected += "Longer than A  Longer than B  Longer than C\n"
	expected += "short          short          short"

	if output != expected {
		printableProof := fmt.Sprintf("\nGot:      %+q", output)
		printableProof += fmt.Sprintf("\nExpected: %+q", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

func TestColumnWidthCalculatorNonASCII(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"⌘⌘⌘⌘⌘⌘⌘⌘ | Longer than B | Longer than C",
		"short | short | short",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B       Column C\n"
	expected += "⌘⌘⌘⌘⌘⌘⌘⌘  Longer than B  Longer than C\n"
	expected += "short     short          short"

	if output != expected {
		printableProof := fmt.Sprintf("\nGot:      %+q", output)
		printableProof += fmt.Sprintf("\nExpected: %+q", expected)
		t.Fatalf("\n%s", printableProof)
	}
}

func BenchmarkColumnWidthCalculator(b *testing.B) {
	// Generate the input
	input := []string{
		"UUID A | UUID B | UUID C | Column D | Column E",
	}

	format := "%s|%s|%s|%s"
	short := "short"

	uuid := func() string {
		buf := make([]byte, 16)
		if _, err := crand.Read(buf); err != nil {
			panic(fmt.Errorf("failed to read random bytes: %v", err))
		}

		return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
			buf[0:4],
			buf[4:6],
			buf[6:8],
			buf[8:10],
			buf[10:16])
	}

	for i := 0; i < 1000; i++ {
		l := fmt.Sprintf(format, uuid()[:8], uuid()[:12], uuid(), short, short)
		input = append(input, l)
	}

	config := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Format(input, config)
	}
}

func TestVariedInputSpacing(t *testing.T) {
	input := []string{
		"Column A       |Column B|    Column C",
		"x|y|          z",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestVariedInputSpacing_NoTrim(t *testing.T) {
	input := []string{
		"Column A|Column B|Column C",
		"x|y|  z",
	}

	config := DefaultConfig()
	config.NoTrim = true
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y           z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestUnmatchedColumnCounts(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"Value A | Value B",
		"Value A | Value B | Value C | Value D",
	}

	config := DefaultConfig()
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "Value A   Value B\n"
	expected += "Value A   Value B   Value C   Value D"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternateDelimiter(t *testing.T) {
	input := []string{
		"Column | A % Column | B % Column | C",
		"Value A % Value B % Value C",
	}

	config := DefaultConfig()
	config.Delim = "%"
	output := Format(input, config)

	expected := "Column | A  Column | B  Column | C\n"
	expected += "Value A     Value B     Value C"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternateSpacingString(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	config.Glue = "    "
	output := Format(input, config)

	expected := "Column A    Column B    Column C\n"
	expected += "x           y           z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestSimpleFormat(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	output := SimpleFormat(input)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestAlternatePrefixString(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := DefaultConfig()
	config.Prefix = "  "
	output := Format(input, config)

	expected := "  Column A  Column B  Column C\n"
	expected += "  x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyFieldReplacement(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | | z",
	}

	config := DefaultConfig()
	config.Empty = "<none>"
	output := Format(input, config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         <none>    z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestEmptyConfigValues(t *testing.T) {
	input := []string{
		"Column A | Column B | Column C",
		"x | y | z",
	}

	config := Config{}
	output := Format(input, &config)

	expected := "Column A  Column B  Column C\n"
	expected += "x         y         z"

	if output != expected {
		t.Fatalf("\nexpected:\n%s\n\ngot:\n%s", expected, output)
	}
}

func TestMergeConfig(t *testing.T) {
	for _, tc := range []struct {
		desc    string
		configA *Config
		configB *Config
		expect  *Config
	}{
		{
			"merges b over a",
			&Config{Delim: "a", Glue: "a", Prefix: "a", Empty: "a"},
			&Config{Delim: "b", Glue: "b", Prefix: "b", Empty: "b"},
			&Config{Delim: "b", Glue: "b", Prefix: "b", Empty: "b"},
		},
		{
			"merges only non-empty config values",
			&Config{Delim: "a", Glue: "a", Prefix: "a", Empty: "a"},
			&Config{Delim: "b", Prefix: "b"},
			&Config{Delim: "b", Glue: "a", Prefix: "b", Empty: "a"},
		},
		{
			"takes b if a is nil",
			nil,
			&Config{Delim: "b", Glue: "b", Prefix: "b", Empty: "b"},
			&Config{Delim: "b", Glue: "b", Prefix: "b", Empty: "b"},
		},
		{
			"takes a if b is nil",
			&Config{Delim: "a", Glue: "a", Prefix: "a", Empty: "a"},
			nil,
			&Config{Delim: "a", Glue: "a", Prefix: "a", Empty: "a"},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			m := MergeConfig(tc.configA, tc.configB)
			if !reflect.DeepEqual(m, tc.expect) {
				t.Fatalf("\nexpect:\n%#v\n\nactual:\n%#v", tc.expect, m)
			}
		})
	}
}
