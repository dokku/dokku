package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
)

// writes the ip to disk
func main() {
	flag.Parse()
	appName := flag.Arg(1)
	procType := flag.Arg(2)
	containerIndex := flag.Arg(3)
	ip := flag.Arg(4)

	if appName == "" {
		common.LogFail("Please specify an app to run the command on")
	}
	err := common.VerifyAppName(appName)
	if err != nil {
		common.LogFail(err.Error())
	}

	appRoot := strings.Join([]string{common.MustGetEnv("DOKKU_ROOT"), appName}, "/")
	filename := fmt.Sprintf("%v/IP.%v.%v", appRoot, procType, containerIndex)
	f, err := os.Create(filename)
	if err != nil {
		common.LogFail(err.Error())
	}
	defer f.Close()

	ipBytes := []byte(ip)
	_, err = f.Write(ipBytes)
	if err != nil {
		common.LogFail(err.Error())
	}
}
