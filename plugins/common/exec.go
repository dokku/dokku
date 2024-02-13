package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"context"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/fatih/color"
)

// ExecCommandInput is the input for the ExecCommand function
type ExecCommandInput struct {
	// Command is the command to execute
	Command string

	// Args are the arguments to pass to the command
	Args []string

	// DisableStdioBuffer disables the stdio buffer
	DisableStdioBuffer bool

	// Env is the environment variables to pass to the command
	Env map[string]string

	// Stdin is the stdin of the command
	Stdin io.Reader

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs
	StreamStdio bool

	// StreamStdout prints stdout directly to os.Stdout as the command runs.
	StreamStdout bool

	// StreamStderr prints stderr directly to os.Stderr as the command runs.
	StreamStderr bool

	// StdoutWriter is the writer to write stdout to
	StdoutWriter io.Writer

	// StderrWriter is the writer to write stderr to
	StderrWriter io.Writer

	// Sudo runs the command with sudo -n -u root
	Sudo bool
}

// ExecCommandResponse is the response for the ExecCommand function
type ExecCommandResponse struct {
	// Stdout is the stdout of the command
	Stdout string

	// Stderr is the stderr of the command
	Stderr string

	// ExitCode is the exit code of the command
	ExitCode int

	// Cancelled is whether the command was cancelled
	Cancelled bool
}

// StdoutContents returns the trimmed stdout of the command
func (ecr ExecCommandResponse) StdoutContents() string {
	return strings.TrimSpace(ecr.Stdout)
}

// StderrContents returns the trimmed stderr of the command
func (ecr ExecCommandResponse) StderrContents() string {
	return strings.TrimSpace(ecr.Stderr)
}

// CallExecCommand executes a command on the local host
func CallExecCommand(input ExecCommandInput) (ExecCommandResponse, error) {
	ctx := context.Background()
	return CallExecCommandWithContext(ctx, input)
}

// CallExecCommandWithContext executes a command on the local host with the given context
func CallExecCommandWithContext(ctx context.Context, input ExecCommandInput) (ExecCommandResponse, error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-signals
		cancel()
	}()

	// hack: colors do not work natively with io.MultiWriter
	// as it isn't detected as a tty. If the output isn't
	// being captured, then color output can be forced.
	isatty := !color.NoColor
	env := os.Environ()
	if isatty && input.DisableStdioBuffer {
		env = append(env, "FORCE_TTY=1")
	}
	if input.Env != nil {
		for k, v := range input.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	command := input.Command
	commandArgs := input.Args
	if input.Sudo {
		commandArgs = append([]string{"-n", "-u", "root", command}, commandArgs...)
		command = "sudo"
	}

	cmd := execute.ExecTask{
		Command:            command,
		Args:               commandArgs,
		Env:                env,
		DisableStdioBuffer: input.DisableStdioBuffer,
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		cmd.PrintCommand = true
	}

	if input.Stdin != nil {
		cmd.Stdin = input.Stdin
	} else if isatty {
		cmd.Stdin = os.Stdin
	}

	if input.StreamStdio {
		cmd.StreamStdio = true
	}
	if input.StreamStdout {
		cmd.StdOutWriter = os.Stdout
	}
	if input.StreamStderr {
		cmd.StdErrWriter = os.Stderr
	}
	if input.StdoutWriter != nil {
		cmd.StdOutWriter = input.StdoutWriter
	}
	if input.StderrWriter != nil {
		cmd.StdErrWriter = input.StderrWriter
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return ExecCommandResponse{
			Stdout:    res.Stdout,
			Stderr:    res.Stderr,
			ExitCode:  res.ExitCode,
			Cancelled: res.Cancelled,
		}, err
	}

	if res.ExitCode != 0 {
		return ExecCommandResponse{
			Stdout:    res.Stdout,
			Stderr:    res.Stderr,
			ExitCode:  res.ExitCode,
			Cancelled: res.Cancelled,
		}, errors.New(res.Stderr)
	}

	return ExecCommandResponse{
		Stdout:    res.Stdout,
		Stderr:    res.Stderr,
		ExitCode:  res.ExitCode,
		Cancelled: res.Cancelled,
	}, nil
}
