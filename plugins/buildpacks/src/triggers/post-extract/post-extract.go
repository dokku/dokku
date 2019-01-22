package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/dokku/dokku/plugins/common"
)

// writes a .buildpacks file into the app
func main() {
	flag.Parse()
	appName := flag.Arg(0)
	tmpWorkDir := flag.Arg(1)

	buildpacks, err := common.PropertyListGet("buildpacks", appName, "buildpacks")
	if err != nil {
		return
	}

	if len(buildpacks) == 0 {
		return
	}

	buildpacksPath := path.Join(tmpWorkDir, ".buildpacks")
	file, err := os.OpenFile(buildpacksPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}

	w := bufio.NewWriter(file)
	for _, buildpack := range buildpacks {
		fmt.Fprintln(w, buildpack)
	}

	if err = w.Flush(); err != nil {
		common.LogFail(fmt.Sprintf("Error writing .buildpacks file: %s", err.Error()))
		return
	}
	file.Chmod(0600)
}
