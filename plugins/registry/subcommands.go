package registry

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// CommandLogin logs a user into the specified server
func CommandLogin(server string, username string, password string, passwordStdin bool) error {
	if passwordStdin {
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		password = strings.TrimSpace(string(stdin))
	}

	if server == "" {
		return errors.New("Missing server argument")
	}
	if username == "" {
		return errors.New("Missing username argument")
	}
	if password == "" {
		return errors.New("Missing password argument")
	}

	if server == "hub.docker.com" || server == "docker.com" {
		server = "docker.io"
	}

	buffer := bytes.Buffer{}
	buffer.Write([]byte(password + "\n"))

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     common.DockerBin(),
		Args:        []string{"login", "--username", username, "--password-stdin", server},
		Stdin:       &buffer,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to run docker login: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run docker login: %s", result.StderrContents())
	}

	_, err = common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger:     "post-registry-login",
		Args:        []string{server, username},
		StreamStdio: true,
		Env: map[string]string{
			"DOCKER_REGISTRY_PASS": password,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// CommandReport displays a registry report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandSet set or clear a registry property for an app
func CommandSet(appName string, property string, value string) error {
	common.CommandPropertySet("registry", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}
