package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	columnize "github.com/ryanuber/columnize"
)

const (
	helpHeader = `Usage: dokku repo[:COMMAND]

Runs commands that interact with the app's repo

Additional commands:`

	helpContent = `
    repo:gc <app>, Runs 'git gc --aggressive' against the application's repo
    repo:purge-cache <app>, Deletes the contents of the build cache stored in the repository
`
)

func main() {
	flag.Usage = usage
	flag.Parse()

	cmd := flag.Arg(0)
	switch cmd {
	case "repo:help":
		usage()
	case "help":
		fmt.Print(helpContent)
	case "repo:gc":
		gitGC()
	case "repo:purge-cache":
		purgeCache()
	default:
		dokkuNotImplementExitCode, err := strconv.Atoi(os.Getenv("DOKKU_NOT_IMPLEMENTED_EXIT"))
		if err != nil {
			fmt.Println("failed to parse DOKKU_NOT_IMPLEMENTED_EXIT")
			dokkuNotImplementExitCode = 10
		}
		os.Exit(dokkuNotImplementExitCode)
	}
}

func usage() {
	config := columnize.DefaultConfig()
	config.Delim = ","
	config.Prefix = "\t"
	config.Empty = ""
	content := strings.Split(helpContent, "\n")[1:]
	fmt.Println(helpHeader)
	fmt.Println(columnize.Format(content, config))
}
