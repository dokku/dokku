package common

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"context"

	execute "github.com/alexellis/go-execute/v2"
	"github.com/fatih/color"
)

type ExecCommandInput struct {
	Command       string
	Args          []string
	CaptureOutput bool
	Env           map[string]string
	StreamStdio   bool
}

func CallExecCommand(input ExecCommandInput) (execute.ExecResult, error) {
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
	env := os.Environ()
	if isatty && !input.CaptureOutput {
		env = append(env, "FORCE_TTY=1")
	}
	if input.Env != nil {
		for k, v := range input.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmd := execute.ExecTask{
		Command:            input.Command,
		Args:               input.Args,
		Env:                env,
		DisableStdioBuffer: !input.CaptureOutput,
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
