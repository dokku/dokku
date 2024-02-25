package run

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func CommandDefault(appName string, command []string, env map[string]string, noTty bool, tty bool, cronID string) error {
	return callRun(CallRunInput{
		AppName:    appName,
		Command:    command,
		ExecuteEnv: env,
		NoTty:      noTty,
		ForceTty:   tty,
		CronID:     cronID,
	})
}

func CommandDetached(appName string, command []string, env map[string]string, noTty bool, tty bool, cronID string) error {
	if !tty && !noTty {
		noTty = true
	}

	return callRun(CallRunInput{
		AppName:    appName,
		Command:    command,
		ExecuteEnv: env,
		Detach:     true,
		NoTty:      noTty,
		ForceTty:   tty,
		CronID:     cronID,
	})
}

func CommandList(appName string, format string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run-list",
		Args:        []string{scheduler, appName, format},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to list run commands: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to list run commands: %s", result.StderrContents())
	}
	return nil
}

func CommandLogs(appName string, container string, numLines int, quiet bool, tail bool) error {
	if appName == "" && container == "" {
		return fmt.Errorf("No container or app specified")
	}

	if container != "" {
		if len(strings.Split(container, ".")) != 3 {
			return fmt.Errorf("Invalid container name specified: %s", container)
		}

		parts := strings.Split(container, ".")
		if appName != "" && appName != parts[0] {
			return fmt.Errorf("Specified app does not app in container name")
		}

		appName = parts[0]
		if parts[1] != "run" {
			return fmt.Errorf("Specified container must be a run container")
		}
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run-logs",
		Args:        []string{scheduler, appName, container, strconv.FormatBool(tail), strconv.FormatBool(quiet), strconv.FormatInt(int64(numLines), 10)},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to retrieve logs: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to retrieve logs: %s", result.StderrContents())
	}

	return nil
}

func CommandStop(appName string, container string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	scheduler := common.GetAppScheduler(appName)
	result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "scheduler-run-stop",
		Args:        []string{scheduler, appName, container},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to stop run container: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to stop run container: %s", result.StderrContents())
	}

	return nil
}
