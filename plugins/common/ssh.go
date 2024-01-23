package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// SshCommandInput is the input for CallSshCommand
type SshCommandInput struct {
	// Command is the command to execute. This can be the path to an executable
	// or the executable with arguments.
	//
	// Any arguments must be given via Args
	Command string

	// Args are the arguments to pass to the command.
	Args []string

	// CaptureOutput saves any output from in the TaskResult
	CaptureOutput bool

	// Env is a list of environment variables to add to the current environment
	Env map[string]string

	// AllowUknownHosts allows connecting to hosts with unknown host keys
	AllowUknownHosts bool

	// RemoteHost is the remote host to connect to
	RemoteHost string

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// Sudo runs the command with sudo -n -u root
	Sudo bool
}

// CallSshCommand executes a command on a remote host via ssh
func CallSshCommand(input SshCommandInput) (SshResult, error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals
		cancel()
	}()

	// hack: colors do not work natively with io.MultiWriter
	// as it isn't detected as a tty. If the output isn't
	// being captured, then color output can be forced.
	isatty := !color.NoColor
	env := []string{}
	if input.Env != nil {
		for k, v := range input.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	u, err := url.Parse(input.RemoteHost)
	if err != nil {
		return SshResult{}, fmt.Errorf("failed to parse remote host: %w", err)
	}
	if u.Scheme == "" {
		return SshResult{}, fmt.Errorf("missing remote host ssh scheme in remote host: %s", input.RemoteHost)
	}
	if u.Scheme != "ssh" {
		return SshResult{}, fmt.Errorf("invalid remote host scheme: %s", u.Scheme)
	}

	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if pass, ok := u.User.Password(); ok {
			password = pass
		}
	}
	if username == "" {
		username = os.Getenv("USER")
	}

	portStr := u.Port()
	port := 0
	if portStr != "" {
		portVal, err := strconv.Atoi(portStr)
		if err != nil {
			return SshResult{}, fmt.Errorf("failed to parse port: %w", err)
		}
		port = portVal
	}
	if port == 0 {
		port = 22
	}

	sshKeyPath := filepath.Join(os.Getenv("DOKKU_ROOT"), ".ssh/id_ed25519")
	if !FileExists(sshKeyPath) {
		sshKeyPath = filepath.Join(os.Getenv("DOKKU_ROOT"), ".ssh/id_rsa")
	}
	if !FileExists(sshKeyPath) {
		return SshResult{}, errors.New("ssh key not found at ~/.ssh/id_ed25519 or ~/.ssh/id_rsa")
	}

	cmd := SshTask{
		Command:            input.Command,
		Args:               input.Args,
		Env:                env,
		DisableStdioBuffer: !input.CaptureOutput,
		AllowUknownHosts:   input.AllowUknownHosts,
		Hostname:           u.Hostname(),
		Port:               uint(port),
		Username:           username,
		Password:           password,
		SshKeyPath:         sshKeyPath,
		Sudo:               input.Sudo,
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		cmd.PrintCommand = true
	}

	if isatty {
		cmd.Stdin = os.Stdin
	}

	if input.StreamStdio {
		cmd.StreamStdio = true
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return res, err
	}

	if res.ExitCode != 0 {
		return res, errors.New(res.Stderr)
	}

	return res, nil
}

// SshTask is a task for executing a command on a remote host via ssh
type SshTask struct {
	// Command is the command to execute. This can be the path to an executable
	// or the executable with arguments.
	//
	// Any arguments must be given via Args
	Command string

	// Args are the arguments to pass to the command.
	Args []string

	// Shell run the command in a bash shell.
	// Note that the system must have `bash` installed in the PATH or in /bin/bash
	Shell bool

	// Env is a list of environment variables to add to the current environment
	Env []string

	// Stdin connect a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// PrintCommand prints the command before executing
	PrintCommand bool
	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// DisableStdioBuffer prevents any output from being saved in the
	// TaskResult, which is useful for when the result is very large, or
	// when you want to stream the output to another writer exclusively.
	DisableStdioBuffer bool

	// StdoutWriter when set will receive a copy of stdout from the command
	StdOutWriter io.Writer

	// StderrWriter when set will receive a copy of stderr from the command
	StdErrWriter io.Writer

	// AllowUknownHosts allows connecting to hosts with unknown host keys
	AllowUknownHosts bool

	// Hostname is the hostname to connect to
	Hostname string

	// Port is the port to connect to
	Port uint

	// Username is the username to connect with
	Username string

	// Password is the password to connect with
	Password string

	// SshKeyPath is the path to the ssh key to use
	SshKeyPath string

	// Sudo runs the command with sudo -n -u root
	Sudo bool
}

// SshResult is the result of executing a command on a remote host via ssh
type SshResult struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Cancelled bool
}

// Execute runs the task
func (task SshTask) Execute(ctx context.Context) (SshResult, error) {
	if task.Command == "" {
		return SshResult{}, errors.New("command is required")
	}
	if task.Hostname == "" {
		return SshResult{}, errors.New("hostname is required")
	}
	if task.SshKeyPath == "" {
		return SshResult{}, errors.New("ssh key path is required")
	}
	if task.Username == "" {
		return SshResult{}, errors.New("username is required")
	}
	if task.Port == 0 {
		task.Port = 22
	}

	if err := TouchFile(filepath.Join(os.Getenv("DOKKU_ROOT"), ".ssh", "known_hosts")); err != nil {
		return SshResult{}, fmt.Errorf("failed to touch known_hosts file: %w", err)
	}

	auth, err := goph.Key(task.SshKeyPath, "")
	if err != nil {
		return SshResult{}, fmt.Errorf("failed to load ssh key: %w", err)
	}

	callback, err := goph.DefaultKnownHosts()
	if err != nil {
		return SshResult{}, fmt.Errorf("failed to load known hosts: %w", err)
	}

	if task.AllowUknownHosts {
		callback = ssh.InsecureIgnoreHostKey()
	}

	connectionConf := goph.Config{
		User:     task.Username,
		Addr:     task.Hostname,
		Port:     task.Port,
		Timeout:  goph.DefaultTimeout,
		Callback: callback,
		Auth:     auth,
	}
	if task.Password != "" {
		connectionConf.Auth = goph.Password(task.Password)
	}

	client, err := goph.NewConn(&connectionConf)
	if err != nil {
		return SshResult{}, fmt.Errorf("failed to create ssh client: %w", err)
	}
	defer client.Close()

	// don't try to run if the context is already cancelled
	if ctx.Err() != nil {
		return SshResult{
			// the exec package returns -1 for cancelled commands
			ExitCode:  -1,
			Cancelled: ctx.Err() == context.Canceled,
		}, ctx.Err()
	}

	var command string
	var commandArgs []string
	if task.Shell {
		command = "bash"
		if len(task.Args) == 0 {
			// use Split and Join to remove any extra whitespace?
			startArgs := strings.Split(task.Command, " ")
			script := strings.Join(startArgs, " ")
			commandArgs = append([]string{"-c"}, script)
		} else {
			script := strings.Join(task.Args, " ")
			commandArgs = append([]string{"-c"}, fmt.Sprintf("%s %s", task.Command, script))
		}
	} else {
		command = task.Command
		commandArgs = task.Args
	}

	if task.Sudo {
		commandArgs = append([]string{"-n", "-u", "root", command}, commandArgs...)
		command = "sudo"
	}

	if task.PrintCommand {
		LogDebug(fmt.Sprintf("ssh %s@%s %s %v", task.Username, task.Hostname, command, commandArgs))
	}

	cmd, err := client.CommandContext(ctx, command, commandArgs...)
	if err != nil {
		return SshResult{}, fmt.Errorf("failed to create ssh command: %w", err)
	}

	if len(task.Env) > 0 {
		overrides := map[string]bool{}
		for _, env := range task.Env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}
	}

	if task.Stdin != nil {
		cmd.Stdin = task.Stdin
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters []io.Writer
	var stderrWriters []io.Writer

	if !task.DisableStdioBuffer {
		stdoutWriters = append(stdoutWriters, &stdoutBuff)
		stderrWriters = append(stderrWriters, &stderrBuff)
	}

	if task.StreamStdio {
		stdoutWriters = append(stdoutWriters, os.Stdout)
		stderrWriters = append(stderrWriters, os.Stderr)
	}

	if task.StdOutWriter != nil {
		stdoutWriters = append(stdoutWriters, task.StdOutWriter)
	}
	if task.StdErrWriter != nil {
		stderrWriters = append(stderrWriters, task.StdErrWriter)
	}

	cmd.Stdout = io.MultiWriter(stdoutWriters...)
	cmd.Stderr = io.MultiWriter(stderrWriters...)

	startErr := cmd.Start()
	if startErr != nil {
		return SshResult{}, startErr
	}

	exitCode := 0
	execErr := cmd.Wait()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return SshResult{
		Stdout:    stdoutBuff.String(),
		Stderr:    stderrBuff.String(),
		ExitCode:  exitCode,
		Cancelled: ctx.Err() == context.Canceled,
	}, ctx.Err()
}
