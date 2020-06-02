package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku config[:COMMAND]

Manage global and app-specific config vars

Additional commands:`

	helpContent = `
    config (<app>|--global), Pretty-print an app or global environment
    config:bundle [--merged] (<app>|--global), Bundle environment into tarfile
    config:clear [--no-restart] (<app>|--global), Clears environment variables
    config:export [--format=FORMAT] [--merged] (<app>|--global), Export a global or app environment
    config:get [--quoted] (<app>|--global) KEY, Display a global or app-specific config value
    config:keys [--merged] (<app>|--global), Show keys set in environment
    config:show [--merged] (<app>|--global), Show keys set in environment
    config:set [--encoded] [--no-restart] (<app>|--global) KEY1=VALUE1 [KEY2=VALUE2 ...], Set one or more config vars
    config:unset [--no-restart] (<app>|--global) KEY1 [KEY2 ...], Unset one or more config vars
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "config":
		common.LogWarn("Deprecated: Use the 'config:show' command instead")
		args := flag.NewFlagSet("config:show", flag.ExitOnError)
		global := args.Bool("global", false, "--global: use the global environment")
		shell := args.Bool("shell", false, "--shell: in a single-line for usage in command-line utilities [deprecated]")
		export := args.Bool("export", false, "--export: print the env as eval-compatible exports [deprecated]")
		merged := args.Bool("merged", false, "--merged: display the app's environment merged with the global environment")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err := config.CommandShow(appName, *global, *merged, *shell, *export)
		if err != nil {
			common.LogFail(err.Error())
		}
	case "config:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    config, Manage global and app-specific config vars\n")
		}
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to retrieve DOKKU_NOT_IMPLEMENTED_EXIT environment variable")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "    "
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}
