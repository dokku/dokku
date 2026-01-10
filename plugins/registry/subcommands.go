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
func CommandLogin(appName string, server string, username string, password string, passwordStdin bool) error {
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

	env := map[string]string{}
	if appName != "" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
		configDir := GetAppRegistryConfigDir(appName)
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("Unable to create registry config directory: %w", err)
		}
		env["DOCKER_CONFIG"] = GetAppRegistryConfigDir(appName)
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"login", "--username", username, "--password-stdin", server},
		Env:     env,
		Stdin:   &buffer,
	})
	if err != nil {
		return fmt.Errorf("Unable to run docker login: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run docker login: %s", result.StderrContents())
	}

	if appName != "" {
		common.LogWarn(fmt.Sprintf("Login Succeeded for %s", appName))
	} else {
		common.LogWarn("Login Succeeded for global registry")
	}

	// todo: change the signature of the trigger to include the app name
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

// CommandLogout logs a user out from the specified server
func CommandLogout(appName string, server string) error {
	if server == "" {
		return errors.New("Missing server argument")
	}

	if server == "hub.docker.com" || server == "docker.com" {
		server = "docker.io"
	}

	env := map[string]string{}
	if appName != "" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
		configDir := GetAppRegistryConfigDir(appName)
		if !common.DirectoryExists(configDir) {
			return fmt.Errorf("No registry credentials found for app %s", appName)
		}
		env["DOCKER_CONFIG"] = GetAppRegistryConfigDir(appName)
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: common.DockerBin(),
		Args:    []string{"logout", server},
		Env:     env,
	})
	if err != nil {
		return fmt.Errorf("Unable to run docker logout: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Unable to run docker logout: %s", result.StderrContents())
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
