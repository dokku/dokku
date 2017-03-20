/*
Package go-sh is intented to make shell call with golang more easily.
Some usage is more similar to os/exec, eg: Run(), Output(), Command(name, args...)

But with these similar function, pipe is added in and this package also got shell-session support.

Why I love golang so much, because the usage of golang is simple, but the power is unlimited. I want to make this pakcage got the sample style like golang.

	// just like os/exec
	sh.Command("echo", "hello").Run()

	// support pipe
	sh.Command("echo", "hello").Command("wc", "-c").Run()

	// create a session to store dir and env
	sh.NewSession().SetDir("/").Command("pwd")

	// shell buildin command - "test"
	sh.Test("dir", "mydir")

	// like shell call: (cd /; pwd)
	sh.Command("pwd", sh.Dir("/")) same with sh.Command(sh.Dir("/"), "pwd")

	// output to json and xml easily
	v := map[string] int {}
	err = sh.Command("echo", `{"number": 1}`).UnmarshalJSON(&v)
*/
package sh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/codegangsta/inject"
)

type Dir string

type Session struct {
	inj     inject.Injector
	alias   map[string][]string
	cmds    []*exec.Cmd
	dir     Dir
	started bool
	Env     map[string]string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	ShowCMD bool // enable for debug
	timeout time.Duration
}

func (s *Session) writePrompt(args ...interface{}) {
	var ps1 = fmt.Sprintf("[golang-sh]$")
	args = append([]interface{}{ps1}, args...)
	fmt.Fprintln(s.Stderr, args...)
}

func NewSession() *Session {
	env := make(map[string]string)
	for _, key := range []string{"PATH"} {
		env[key] = os.Getenv(key)
	}
	s := &Session{
		inj:    inject.New(),
		alias:  make(map[string][]string),
		dir:    Dir(""),
		Stdin:  strings.NewReader(""),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    env,
	}
	return s
}

func InteractiveSession() *Session {
	s := NewSession()
	s.SetStdin(os.Stdin)
	return s
}

func Command(name string, a ...interface{}) *Session {
	s := NewSession()
	return s.Command(name, a...)
}

func Echo(in string) *Session {
	s := NewSession()
	return s.SetInput(in)
}

func (s *Session) Alias(alias, cmd string, args ...string) {
	v := []string{cmd}
	v = append(v, args...)
	s.alias[alias] = v
}

func (s *Session) Command(name string, a ...interface{}) *Session {
	var args = make([]string, 0)
	var sType = reflect.TypeOf("")

	// init cmd, args, dir, envs
	// if not init, program may panic
	s.inj.Map(name).Map(args).Map(s.dir).Map(map[string]string{})
	for _, v := range a {
		switch reflect.TypeOf(v) {
		case sType:
			args = append(args, v.(string))
		default:
			s.inj.Map(v)
		}
	}
	if len(args) != 0 {
		s.inj.Map(args)
	}
	s.inj.Invoke(s.appendCmd)
	return s
}

// combine Command and Run
func (s *Session) Call(name string, a ...interface{}) error {
	return s.Command(name, a...).Run()
}

/*
func (s *Session) Exec(cmd string, args ...string) error {
	return s.Call(cmd, args)
}
*/

func (s *Session) SetEnv(key, value string) *Session {
	s.Env[key] = value
	return s
}

func (s *Session) SetDir(dir string) *Session {
	s.dir = Dir(dir)
	return s
}

func (s *Session) SetInput(in string) *Session {
	s.Stdin = strings.NewReader(in)
	return s
}

func (s *Session) SetStdin(r io.Reader) *Session {
	s.Stdin = r
	return s
}

func (s *Session) SetTimeout(d time.Duration) *Session {
	s.timeout = d
	return s
}

func newEnviron(env map[string]string, inherit bool) []string { //map[string]string {
	environ := make([]string, 0, len(env))
	if inherit {
		for _, line := range os.Environ() {
			for k, _ := range env {
				if strings.HasPrefix(line, k+"=") {
					goto CONTINUE
				}
			}
			environ = append(environ, line)
		CONTINUE:
		}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	return environ
}

func (s *Session) appendCmd(cmd string, args []string, cwd Dir, env map[string]string) {
	if s.started {
		s.started = false
		s.cmds = make([]*exec.Cmd, 0)
	}
	for k, v := range s.Env {
		if _, ok := env[k]; !ok {
			env[k] = v
		}
	}
	environ := newEnviron(s.Env, true) // true: inherit sys-env
	v, ok := s.alias[cmd]
	if ok {
		cmd = v[0]
		args = append(v[1:], args...)
	}
	c := exec.Command(cmd, args...)
	c.Env = environ
	c.Dir = string(cwd)
	s.cmds = append(s.cmds, c)
}
