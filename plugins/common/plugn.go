package common

import (
	execute "github.com/alexellis/go-execute/v2"
)

type PlugnTriggerInput struct {
	Args          []string
	CaptureOutput bool
	Env           map[string]string
	StreamStdio   bool
	Trigger       string
}

func CallPlugnTrigger(input PlugnTriggerInput) (execute.ExecResult, error) {
	args := []string{"trigger", input.Trigger}
	args = append(args, input.Args...)
	return CallExecCommand(ExecCommandInput{
		Command:       "plugn",
		Args:          args,
		CaptureOutput: input.CaptureOutput,
		Env:           input.Env,
		StreamStdio:   input.StreamStdio,
	})
}
