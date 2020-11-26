package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/logs"
)

const (
	helpHeader = `Usage: dokku logs[:COMMAND]

Manage log integration for an app

Additional commands:`

	helpContent = `
    logs [-h] [-t] [-n num] [-q] [-p process] <app>, Display recent log output
    logs:failed [<app>], Shows the last failed deploy logs
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "logs":
		args := flag.NewFlagSet("logs", flag.ExitOnError)
		var num int64
		var ps string
		var tail bool
		var quiet bool
		help := args.Bool("h", false, "-h: print help for the command")
		args.Int64Var(&num, "n", 0, "-n: the number of lines to display")
		args.Int64Var(&num, "num", 0, "--num: the number of lines to display")
		args.StringVar(&ps, "p", "", "-p: only display logs from the given process")
		args.StringVar(&ps, "ps", "", "--p: only display logs from the given process")
		args.BoolVar(&tail, "t", true, "-t: continually stream logs")
		args.BoolVar(&tail, "tail", true, "--tail: continually stream logs")
		args.BoolVar(&quiet, "q", true, "-q: display raw logs without colors, time and names")
		args.BoolVar(&quiet, "quiet", true, "--quiet: display raw logs without colors, time and names")
		args.Parse(flag.Args()[1:])
		if *help {
			usage()
			return
		}
		appName := args.Arg(0)
		err := logs.CommandDefault(appName, num, ps, tail, quiet)
		if err != nil {
			common.LogFail(err.Error())
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
