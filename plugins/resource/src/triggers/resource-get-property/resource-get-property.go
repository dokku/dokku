package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
)

// writes the resource key to stdout for a given app container
func main() {
	flag.Parse()
	appName := flag.Arg(0)
	processType := flag.Arg(1)
	resourceType := flag.Arg(2)
	key := flag.Arg(3)

	value, err := resource.GetResourceValue(appName, processType, resourceType, key)
	if err != nil {
		common.LogFail(err.Error())
	}

	fmt.Fprintln(os.Stdout, value)
}
