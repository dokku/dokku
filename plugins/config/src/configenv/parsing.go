package configenv

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
)

func parseEnv(name string, filename string) (*Env, error) {
	var file *os.File
	var err error
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err = os.Create(filename)
	} else {
		file, err = os.Open(filename)
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseEnvFromReader(name, filename, file)
}

func parseEnvFromReader(name string, filename string, reader io.Reader) (*Env, error) {
	//The file format looks like: `KEY='VALUE'`. Eait pair is terminated with a newline.
	//All characters are valid for VALUE without escaping but one: the single quote.
	//A single quote literal is represented by ending the quoting
	//emitting a \' and resuming the quoting. So for VALUE to be "don't care":
	//`KEY='don'\''t care'`
	//The advantage of this is that it's easily parsed and sourced by bash correctly.

	buffered := bufio.NewReader(reader)
	const (
		StateKey   = iota
		StateValue = iota
	)
	env := &Env{name, map[string]string{}, filename}
	var buffer bytes.Buffer
	var state = StateKey
	var quoted = false
	var escaped = false
	var key = ""
	for char, _, err := buffered.ReadRune(); err == nil; char, _, err = buffered.ReadRune() {
		switch state {
		case StateKey:
			switch char {
			case ' ':
				if buffer.String() == "export" {
					buffer.Truncate(0) //so we can read exportfiles as well as envfiles
				} else if buffer.Len() == 0 {
					continue //leading spaced are allowed but not encouraged
				} else {
					return nil, errors.New("keys cannot have spaces")
				}
			case '=':
				key = buffer.String()
				buffer.Truncate(0)
				state = StateValue
			case '\n':
				if buffer.Len() > 0 {
					return nil, errors.New("keys cannot contain newlines")
				}
			default:
				buffer.WriteRune(char)
			}
		case StateValue:
			switch char {
			case '\'':
				if escaped {
					buffer.WriteRune('\'')
					escaped = false
				} else {
					quoted = !quoted
				}
			case '\n':
				if quoted {
					buffer.WriteRune(char)
				} else {
					state = StateKey
					env.Set(key, buffer.String())
					buffer.Truncate(0)
					key = ""
					escaped = false
					quoted = false
				}
			case '\\':
				if quoted {
					buffer.WriteRune(char)
					escaped = false
				} else {
					escaped = true
				}
			default:
				if !quoted {
					return nil, errors.New("unquoted or unbalanced value for " + key)
				}
				if escaped {
					return nil, errors.New("invalid escape for " + key)
				}
				buffer.WriteRune(char)
			}
		}
	}
	if escaped {
		return nil, errors.New("unterminated escape at end")
	}
	if quoted {
		return nil, errors.New("unterminated quote at end")
	}

	if state == StateValue {
		env.Set(key, buffer.String())
	} else {
		if buffer.Len() > 0 {
			return nil, errors.New("invalid trailing content: '" + buffer.String() + "'")
		}
	}
	return env, nil
}
