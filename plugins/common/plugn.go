package common

import (
	"context"
	"io"
)

// PlugnTriggerInput is the input for CallPlugnTrigger
type PlugnTriggerInput struct {
	// Args are the arguments to pass to the trigger
	Args []string

	// DisableStdioBuffer disables the stdio buffer
	DisableStdioBuffer bool

	// Env is the environment variables to pass to the trigger
	Env map[string]string

	// Stdin is the stdin of the command
	Stdin io.Reader

	// StreamStdio determines whether to stream the stdio of the trigger
	StreamStdio bool

	// StreamStdout prints stdout directly to os.Stdout as the command runs.
	StreamStdout bool

	// StreamStderr prints stderr directly to os.Stderr as the command runs.
	StreamStderr bool

	// Trigger is the trigger to execute
	Trigger string
}

// CallPlugnTrigger executes a trigger via plugn
func CallPlugnTrigger(input PlugnTriggerInput) (ExecCommandResponse, error) {
	return CallPlugnTriggerWithContext(context.Background(), input)
}

// CallPlugnTriggerWithContext executes a trigger via plugn with the given context
func CallPlugnTriggerWithContext(ctx context.Context, input PlugnTriggerInput) (ExecCommandResponse, error) {
	args := []string{"trigger", input.Trigger}
	args = append(args, input.Args...)
	return CallExecCommandWithContext(ctx, ExecCommandInput{
		Command:            "plugn",
		Args:               args,
		DisableStdioBuffer: input.DisableStdioBuffer,
		Env:                input.Env,
		Stdin:              input.Stdin,
		StreamStdio:        input.StreamStdio,
		StreamStdout:       input.StreamStdout,
		StreamStderr:       input.StreamStderr,
	})
}
