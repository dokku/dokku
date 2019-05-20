package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// outputs the process-specific docker options
func main() {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		common.LogFail(err.Error())
	}

	flag.Parse()
	appName := flag.Arg(0)
	processType := flag.Arg(3)

	resources, err := common.PropertyGetAll("resource", appName)
	if err != nil {
		fmt.Print(string(stdin))
		return
	}

	limits := make(map[string]string)
	reservations := make(map[string]string)

	validLimits := map[string]bool{
		"cpu":         true,
		"memory":      true,
		"memory-swap": true,
	}
	validReservations := map[string]bool{
		"memory": true,
	}
	validPrefixes := []string{"_default_.", fmt.Sprintf("%s.", processType)}
	for _, validPrefix := range validPrefixes {
		for key, value := range resources {
			if !strings.HasPrefix(key, validPrefix) {
				continue
			}
			parts := strings.SplitN(strings.TrimPrefix(key, validPrefix), ".", 2)
			if parts[0] == "limit" {
				if !validLimits[parts[1]] {
					continue
				}

				if parts[1] == "cpu" {
					parts[1] = "cpus"
				}

				limits[parts[1]] = value
			}
			if parts[0] == "reserve" {
				if !validReservations[parts[1]] {
					continue
				}

				reservations[parts[1]] = value
			}
		}
	}

	for key, value := range limits {
		if value == "" {
			continue
		}
		fmt.Printf(" --%s=%s ", key, value)
	}

	for key, value := range reservations {
		if value == "" {
			continue
		}
		fmt.Printf(" --%s-reservation=%s ", key, value)
	}

	fmt.Print(string(stdin))
}
