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
	//The file format looks like KEY='VALUE'. Eait pair is terminated with a newline.
	//All characters are valid for VALUE without escaping
	//but one: the single quote and backslash. To represent a single quote emit \'. To represent
	//a literal backslash emit \\. In any other case, the backslash is interpreted literally.
	//KEY='value
	//one'
	//KEY='val\\ue\\'' etc..

	//TODO: We might want to rework this to use shell-style quoting for single quotes
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
					buffer.Truncate(0) //so we can read exportfiles too
				} else {
					return nil, errors.New("Env keys cannot have spaces")
				}
			case '=':
				key = buffer.String()
				buffer.Truncate(0)
				state = StateValue
			case '\n':
				if buffer.Len() > 0 {
					return nil, errors.New("Invalid newline after: " + buffer.String())
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
				if escaped || quoted {
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
				if escaped {
					buffer.WriteRune(char)
					escaped = false
				} else {
					escaped = true
				}
			default:
				if escaped {
					buffer.WriteRune('\\')
					escaped = false
				}
				buffer.WriteRune(char)
			}
		}
	}
	if escaped {
		return nil, errors.New("Unterminated escape")
	}
	if quoted {
		return nil, errors.New("Unterminated quote")
	}

	if state == StateValue {
		env.Set(key, buffer.String())
	} else {
		if buffer.Len() > 0 {
			return nil, errors.New("Invalid trailing content: " + buffer.String())
		}
	}
	return env, nil
}
