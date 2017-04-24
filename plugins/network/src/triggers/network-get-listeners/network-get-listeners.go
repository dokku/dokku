package main

import (
	"path/filepath"
	"flag"
	"fmt"
	"os"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
)

// returns the listeners (host:port combinations) for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(1)

	dokkuRoot := common.MustGetEnv("DOKKU_ROOT")
	appRoot := strings.Join([]string{dokkuRoot, appName}, "/")

	files, _ := filepath.Glob(appRoot + "/IP.web.*")

	var listeners []string
	for _, ipfile := range files {
		portfile := strings.Replace(ipfile, "/IP.web.", "/PORT.web.", 1)
		ipAddress := common.ReadFirstLine(ipfile)
		port := common.ReadFirstLine(portfile)
		listeners = append(listeners, fmt.Sprintf("%s:%s", ipAddress, port))
	}

	fmt.Fprint(os.Stdout, strings.Join(listeners, " "))
}
