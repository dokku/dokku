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
	helpHeader = `Usage: dokku network[:COMMAND]

Manage network settings for an app

Additional commands:`

	helpContent = `
    network:create <network>, Creates an attachable docker network
    network:destroy <network>, Destroys a docker network
    network:exists <network>, Checks if a docker network exists
    network:info <network>, Outputs information about a docker network
    network:list, Lists all docker networks
    network:report [<app>] [<flag>], Displays a network report for one or more apps
    network:rebuild <app>, Rebuilds network settings for an app
    network:rebuildall, Rebuild network settings for all apps
    network:set <app> <property> (<value>), Set or clear a network property for an app
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "network", "network:help":
		usage()
	case "help":
		command := common.NewShellCmd(fmt.Sprintf("ps -o command= %d", os.Getppid()))
		command.ShowOutput = false
		output, err := command.Output()

		if err == nil && strings.Contains(string(output), "--all") {
			fmt.Println(helpContent)
		} else {
			fmt.Print("\n    network, Manage network settings for an app\n")
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
