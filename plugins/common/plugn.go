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

type PlugnTriggerInput struct {
	Args          []string
	CaptureOutput bool
	Env           map[string]string
	StreamStdio   bool
	Trigger       string
}

func CallPlugnTrigger(input PlugnTriggerInput) (execute.ExecResult, error) {
	LogDebug(fmt.Sprintf("plugn trigger %s %v", input.Trigger, input.Args))

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

	args := []string{"trigger", input.Trigger}
	args = append(args, input.Args...)
	cmd := execute.ExecTask{
		Command:            "plugn",
		Args:               args,
		Env:                env,
		DisableStdioBuffer: !input.CaptureOutput,
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
