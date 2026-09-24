package shellwords

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// Split splits input into shell words using the parser directly, so quotes
// delimit words and are stripped from the returned value, but nothing is
// expanded. Parameter expansions, command substitutions, and other
// metacharacters are preserved verbatim. Malformed input, such as an unbalanced
// quote, returns the parser error.
func Split(input string) ([]string, error) {
	fields, _, err := SplitWithEnd(input)
	return fields, err
}

// SplitWithEnd is Split plus the byte offset just past the last word the
// parser consumed. Callers that must not lose data compare the offset against
// the length of the input to detect a tail the parser skipped, such as a shell
// comment.
func SplitWithEnd(input string) ([]string, int, error) {
	parser := syntax.NewParser()
	var fields []string
	end := 0
	for word, err := range parser.WordsSeq(strings.NewReader(input)) {
		if err != nil {
			return nil, 0, err
		}
		fields = append(fields, literalWordValue(input, word.Parts))
		end = int(word.End().Offset())
	}
	return fields, end, nil
}

// literalWordValue reconstructs the unquoted, unexpanded value of a shell word.
// Literal and single-quoted parts contribute their literal text; double-quoted
// parts have their surrounding quotes dropped while their contents stay
// literal; every other part - parameter expansions, command substitutions,
// arithmetic expansions - contributes its original source text unchanged.
// Backslash escapes are removed the same way the shell removes them during
// quote removal so a value such as `Host(\`app\`)` round-trips to `Host(`app`)`.
func literalWordValue(input string, parts []syntax.WordPart) string {
	var sb strings.Builder
	for _, part := range parts {
		switch p := part.(type) {
		case *syntax.Lit:
			sb.WriteString(unquoteBackslashes(p.Value, false))
		case *syntax.SglQuoted:
			sb.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, inner := range p.Parts {
				if lit, ok := inner.(*syntax.Lit); ok {
					sb.WriteString(unquoteBackslashes(lit.Value, true))
					continue
				}
				sb.WriteString(input[inner.Pos().Offset():inner.End().Offset()])
			}
		default:
			sb.WriteString(input[part.Pos().Offset():part.End().Offset()])
		}
	}
	return sb.String()
}

// unquoteBackslashes removes backslashes the way the shell does during quote
// removal. Inside double quotes only \$, \`, \", \\, and an escaped newline
// lose their backslash; unquoted, a backslash escapes any following character.
func unquoteBackslashes(s string, inDblQuotes bool) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			if next == '\n' {
				i++
				continue
			}
			if !inDblQuotes || next == '$' || next == '`' || next == '"' || next == '\\' {
				sb.WriteByte(next)
				i++
				continue
			}
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}
