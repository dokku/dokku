package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/logs"

	stdFlag "flag"

	flag "github.com/spf13/pflag"
)

const (
	helpHeader = `Usage: dokku logs[:COMMAND]

Manage log integration for an app

Additional commands:`

	helpContent = `
    logs [-h] [-t|--tail] [-n|--num num] [-q|--quiet] [-p|--ps process] <app>, Display recent log output
    logs:failed [--all|<app>], Shows the last failed deploy logs
    logs:report [<app>] [<flag>], Displays a logs report for one or more apps
    logs:set [--global|<app>] <key> <value>, Set or clear a logs property for an app
    logs:vector-logs [--num num] [--tail], Display vector log output
    logs:vector-start, Start the vector logging container
    logs:vector-stop, Stop the vector logging container
`
)

func main() {
	stdFlag.Usage = usage
	stdFlag.Parse()

	cmd := stdFlag.Arg(0)
	switch cmd {
	case "logs":
		args := flag.NewFlagSet("logs", flag.ExitOnError)
		help := args.Bool("h", false, "-h: print help for the command")
		num := args.Int64P("num", "n", 100, "the number of lines to display")
		ps := args.StringP("ps", "p", "", "only display logs from the given process")
		tail := args.BoolP("tail", "t", false, "continually stream logs")
		quiet := args.BoolP("quiet", "q", false, "display raw logs without colors, time and names")
		args.Parse(os.Args[2:])
		if *help {
			usage()
			return
		}

		appName := args.Arg(0)
		err := logs.CommandDefault(appName, *num, *ps, *tail, *quiet)
		if err != nil {
			common.LogFailWithError(err)
		}
	case "logs:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    logs, Manage log integration for an app\n")
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
	common.CommandUsage(helpHeader, helpContent)
}
