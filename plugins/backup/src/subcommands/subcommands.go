package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/backup"
	"github.com/dokku/dokku/plugins/common"
	flag "github.com/spf13/pflag"
)

// main dispatches the backup subcommands (export/import).
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "export":
		args := flag.NewFlagSet("backup:export", flag.ExitOnError)
		appNames := args.StringArray("app", []string{}, "--app: an app to export; repeatable")
		serviceSpecs := args.StringArray("service", []string{}, "--service: a service to export as TYPE[:NAME]; repeatable")
		backupDir := args.String("backup-dir", "/tmp", "--backup-dir: directory the backup archive is written to")
		includeStorage := args.Bool("include-storage", false, "--include-storage: bundle persistent storage volume data")
		args.Parse(os.Args[2:])
		err = backup.CommandExport(*appNames, *serviceSpecs, *backupDir, *includeStorage)
	case "import":
		args := flag.NewFlagSet("backup:import", flag.ExitOnError)
		appNames := args.StringArray("app", []string{}, "--app: an app to restore; repeatable")
		serviceSpecs := args.StringArray("service", []string{}, "--service: a service to restore as TYPE[:NAME]; repeatable")
		force := args.Bool("force", false, "--force: replace existing apps/services without confirmation")
		skipInstallPlugins := args.Bool("skip-install-plugins", false, "--skip-install-plugins: do not reinstall third-party plugins recorded in the backup; only report them")
		args.Parse(os.Args[2:])
		backupFile := args.Arg(0)
		if backupFile == "" {
			common.LogFail("Please specify a backup file to import")
		}
		err = backup.CommandImport(backupFile, *appNames, *serviceSpecs, *force, !*skipInstallPlugins)
	default:
		err = fmt.Errorf("Invalid plugin subcommand call: %s", subcommand)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
