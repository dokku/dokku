package run

import (
	"fmt"
	"strconv"

	"github.com/dokku/dokku/plugins/common"
)

// CallRunInput is the input object for the callRun function
type CallRunInput struct {
	// AppName is the name of the app
	AppName string

	// Command is the command to run
	Command []string

	// ExecuteEnv is the environment to execute the trigger in
	ExecuteEnv map[string]string

	// Detach specifies if the command should be run in detached mode
	Detach bool

	// RunEnv is the environment to run the command in
	RunEnv map[string]string

	// NoTty specifies if a pseudo-TTY should not be allocated
	NoTty bool

	// ForceTty specifies if a pseudo-TTY should be forced
	ForceTty bool

	// CronID is the cron job id
	CronID string
}

// callRun runs a command in a container
func callRun(input CallRunInput) error {
	if err := common.VerifyAppName(input.AppName); err != nil {
		return err
	}

	env := map[string]string{}
	env["DOKKU_RM_CONTAINER"] = "1"
	if input.Detach {
		env["DOKKU_DETACH_CONTAINER"] = "1"
	}

	if input.NoTty {
		env["DOKKU_DISABLE_TTY"] = "1"
	}
	if input.ForceTty {
		env["DOKKU_FORCE_TTY"] = "1"
	}
	if input.CronID != "" {
		env["CRON_ID"] = input.CronID
	}

	if input.NoTty && input.ForceTty {
		return fmt.Errorf("Cannot specify both --tty and --no-tty")
	}

	runEnv := []string{}
	for key, value := range input.RunEnv {
		runEnv = append(runEnv, fmt.Sprintf("%s=%s", key, value))
	}

	runEnvCount := len(runEnv)

	scheduler := common.GetAppScheduler(input.AppName)
	args := []string{scheduler, input.AppName}
	args = append(args, strconv.FormatInt(int64(runEnvCount), 10))
	args = append(args, runEnv...)
	args = append(args, "--")
	args = append(args, input.Command...)

	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run",
		Args:        args,
		Env:         env,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to run command: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run command: %s", result.StderrContents())
	}
	return nil
}
