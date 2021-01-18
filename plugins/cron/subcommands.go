package cron

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
	"github.com/ryanuber/columnize"
)

// CommandList lists all scheduled cron tasks for a given app
func CommandList(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	entries, err := fetchCronEntries(appName)
	if err != nil {
		return err
	}

	output := []string{"ID | Schedule | Command"}
	for _, entry := range entries {
		output = append(output, fmt.Sprintf("%s | %s | %s", entry.ID, entry.Schedule, entry.Command))
	}
	result := columnize.SimpleFormat(output)
	fmt.Println(result)

	return nil
}

// CommandReport displays a cron report for one or more apps
func CommandReport(appName string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, infoFlag)
}
