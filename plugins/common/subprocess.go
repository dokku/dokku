package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/codeskyblue/go-sh"
)

// ShellCmd represents a shell command to be run for dokku
type ShellCmd struct {
	Env           map[string]string
	Command       *exec.Cmd
	CommandString string
	Args          []string
	ShowOutput    bool
}

// NewShellCmd returns a new ShellCmd struct
func NewShellCmd(command string) *ShellCmd {
	items := strings.Split(command, " ")
	cmd := items[0]
	args := items[1:]
	return NewShellCmdWithArgs(cmd, args...)
}

// NewShellCmdWithArgs returns a new ShellCmd struct
func NewShellCmdWithArgs(cmd string, args ...string) *ShellCmd {
	commandString := strings.Join(append([]string{cmd}, args...), " ")

	return &ShellCmd{
		Command:       exec.Command(cmd, args...),
		CommandString: commandString,
		Args:          args,
		ShowOutput:    true,
	}
}

func (sc *ShellCmd) setup() {
	env := os.Environ()
	for k, v := range sc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	sc.Command.Env = env
	if sc.ShowOutput {
		sc.Command.Stdout = os.Stdout
		sc.Command.Stderr = os.Stderr
	}
}

// Execute is a lightweight wrapper around exec.Command
func (sc *ShellCmd) Execute() bool {
	sc.setup()

	if err := sc.Command.Run(); err != nil {
		return false
	}
	return true
}

// Output is a lightweight wrapper around exec.Command.Output()
func (sc *ShellCmd) Output() ([]byte, error) {
	sc.setup()
	return sc.Command.Output()
}

// CombinedOutput is a lightweight wrapper around exec.Command.CombinedOutput()
func (sc *ShellCmd) CombinedOutput() ([]byte, error) {
	sc.setup()
	return sc.Command.CombinedOutput()
}

// PlugnTrigger fire the given plugn trigger with the given args
func PlugnTrigger(triggerName string, args ...string) error {
	return PlugnTriggerSetup(triggerName, args...).Run()
}

// PlugnTriggerOutput fire the given plugn trigger with the given args
func PlugnTriggerOutput(triggerName string, args ...string) ([]byte, error) {
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	out, err := PlugnTriggerSetup(triggerName, args...).Output()
	w.Close()

	readStderr, _ := ioutil.ReadAll(r)
	os.Stderr = rescueStderr

	var stderr error
	if err != nil {
		stderr = fmt.Errorf(string(readStderr[:]))
	}

	return out, stderr
}

// PlugnTriggerSetup sets up a plugn trigger call
func PlugnTriggerSetup(triggerName string, args ...string) *sh.Session {
	shellArgs := make([]interface{}, len(args)+2)
	shellArgs[0] = "trigger"
	shellArgs[1] = triggerName
	for i, arg := range args {
		shellArgs[i+2] = arg
	}
	return sh.Command("plugn", shellArgs...)
}
