package scheduler_k3s

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	resty "github.com/go-resty/resty/v2"
)

// CommandInitialize initializes a k3s cluster on the local server
func CommandInitialize() error {
	client := resty.New()
	resp, err := client.R().
		Get("https://get.k3s.io")
	if err != nil {
		return fmt.Errorf("Unable to download k3s installer: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("Missing response from k3s installer download: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Invalid status code for k3s installer script: %d", resp.StatusCode())
	}

	f, err := os.CreateTemp("", "sample")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file for k3s installer: %w", err)
	}
	defer os.Remove(f.Name())

	if err := f.Close(); err != nil {
		return fmt.Errorf("Unable to close k3s installer file: %w", err)
	}

	err = common.WriteSliceToFile(common.WriteSliceToFileInput{
		Filename: f.Name(),
		Lines:    strings.Split(resp.String(), "\n"),
		Mode:     os.FileMode(755),
	})
	if err != nil {
		return fmt.Errorf("Unable to write k3s installer to file: %w", err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return fmt.Errorf("Unable to get k3s installer file size: %w", err)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Invalid k3s installer filesize")
	}

	// todo: allow this to be passed as an option or environment variable
	token := "password"
	if err := CommandSet("--global", "token", token); err != nil {
		return fmt.Errorf("Unable to set k3s token: %w", err)
	}

	installerCmd := common.NewShellCmd(strings.Join([]string{
		f.Name(),
		// initialize the cluster
		"--cluster-init",
		// allow access for the dokku user
		"--write-kubeconfig-mode", "0644",
		// specify a token
		"--token", token,
	}, " "))
	if !installerCmd.Execute() {
		return fmt.Errorf("Error installing k3s: %w", installerCmd.ExitError)
	}

	if err := common.TouchFile("/etc/rancher/k3s/registries.yaml"); err != nil {
		return fmt.Errorf("Error creating initial registries.yaml file")
	}

	registryAclCmd := common.NewShellCmd(strings.Join([]string{
		"setfacl",
		"-m",
		"user:dokku:rwx",
		"/etc/rancher/k3s/registries.yaml",
	}, " "))
	if !registryAclCmd.Execute() {
		return fmt.Errorf("Error updating acls on k3s registries.yaml file: %w", registryAclCmd.ExitError)
	}

	return nil
}

// CommandReport displays a scheduler-k3s report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
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

// CommandSet set or clear a scheduler-k3s property for an app
func CommandSet(appName string, property string, value string) error {
	common.CommandPropertySet("scheduler-k3s", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}
