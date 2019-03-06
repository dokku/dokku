package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku buildpacks[:COMMAND]

Manages buildpacks settings for an app

Additional commands:`

	helpContent = `
    buildpacks:add [--index 1] <app> <buildpack>, Add new app buildpack while inserting into list of buildpacks if necessary
    buildpacks:clear <app>, Clear all buildpacks set on the app
    buildpacks:list <app>, List all buildpacks for an app
    buildpacks:remove <app> <buildpack>, Remove a buildpack set on the app
    buildpacks:report [<app>] [<flag>], Displays a buildpack report for one or more apps
    buildpacks:set [--index 1] <app> <buildpack>, Set new app buildpack at a given position defaulting to the first buildpack if no index is specified
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "buildpacks", "buildpacks:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    buildpacks, Manages buildpack settings for an app\n")
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
