package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

// Start is a wrapper around exec.Command.Start()
func (sc *ShellCmd) Start() error {
	sc.setup()

	return sc.Command.Start()
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
	LogDebug(fmt.Sprintf("plugn trigger %s %v", triggerName, args))
	return PlugnTriggerSetup(triggerName, args...).Run()
}

// PlugnTriggerOutput fire the given plugn trigger with the given args
func PlugnTriggerOutput(triggerName string, args ...string) ([]byte, error) {
	LogDebug(fmt.Sprintf("plugn trigger %s %v", triggerName, args))
	rE, wE, _ := os.Pipe()
	rO, wO, _ := os.Pipe()
	session := PlugnTriggerSetup(triggerName, args...)
	session.Stderr = wE
	session.Stdout = wO
	err := session.Run()
	wE.Close()
	wO.Close()

	readStderr, _ := ioutil.ReadAll(rE)
	readStdout, _ := ioutil.ReadAll(rO)

	stderr := string(readStderr[:])
	if err != nil {
		err = fmt.Errorf(stderr)
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		for _, line := range strings.Split(stderr, "\n") {
			LogDebug(fmt.Sprintf("plugn trigger %s stderr: %s", triggerName, line))
		}
		for _, line := range strings.Split(string(readStdout[:]), "\n") {
			LogDebug(fmt.Sprintf("plugn trigger %s stdout: %s", triggerName, line))
		}
	}

	return readStdout, err
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

// PlugnTriggerExists returns whether a plugin trigger exists (ignoring the existence of any within the 20_events plugin)
func PlugnTriggerExists(triggerName string) bool {
	pluginPath := MustGetEnv("PLUGIN_ENABLED_PATH")
	glob := filepath.Join(pluginPath, "*", triggerName)
	exists := false
	files, _ := filepath.Glob(glob)
	for _, file := range files {
		plugin := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(file, "/"+triggerName), pluginPath), "/")
		if plugin != "20_events" {
			exists = true
			break
		}
	}
	return exists
}
