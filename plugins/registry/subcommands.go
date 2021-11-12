package registry

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/apps"
	"github.com/dokku/dokku/plugins/common"
)

// CommandLogin logs a user into the specified server
func CommandLogin(server string, username string, password string, passwordStdin bool) error {
	if passwordStdin {
		stdin, err := ioutil.ReadAll(os.Stdin)
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

	command := []string{
		common.DockerBin(),
		"login",
		"--username",
		username,
		"--password-stdin",
		server,
	}

	buffer := bytes.Buffer{}
	buffer.Write([]byte(password + "\n"))

	loginCmd := common.NewShellCmd(strings.Join(command, " "))
	loginCmd.Command.Stdin = &buffer
	if !loginCmd.Execute() {
		return errors.New("Failed to log into registry")
	}

	return nil
}

// CommandReport displays a registry report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := apps.DokkuApps()
		if err != nil {
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
