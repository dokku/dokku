package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

func strToInt(s string, defaultValue int, allowEmpty bool) int {
	if s == "" {
		if !allowEmpty {
			os.Exit(1)
		}
		return defaultValue
	}

	index, err := strconv.Atoi(s)
	if err != nil {
		os.Exit(1)
	}

	return index
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)
	pluginName := flag.Arg(1)

	switch cmd {
	case "rpush":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := flag.Arg(4)
		index := strToInt(flag.Arg(5), 0, true)
		err := common.PropertyListAdd(pluginName, appName, property, value, index)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "llen":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		length, err := common.PropertyListLength(pluginName, appName, property)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Printf("%v\n", length)
	case "lrange":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		lines, err := common.PropertyListGet(pluginName, appName, property)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		for _, line := range lines {
			fmt.Printf("%v\n", line)
		}
	case "lindex":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		index := strToInt(flag.Arg(4), 0, false)
		propertyValue, err := common.PropertyListGetByIndex(pluginName, appName, property, index)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		fmt.Printf("%v\n", propertyValue)
	case "lismember":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := flag.Arg(4)
		_, err := common.PropertyListGetByValue(pluginName, appName, property, value)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "lrem":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := flag.Arg(4)
		err := common.PropertyListRemove(pluginName, appName, property, value)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "lrem-by-prefix":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		prefix := flag.Arg(4)
		err := common.PropertyListRemoveByPrefix(pluginName, appName, property, prefix)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "lset":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := flag.Arg(4)
		index := strToInt(flag.Arg(5), 0, false)
		err := common.PropertyListSet(pluginName, appName, property, value, index)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "set":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := flag.Arg(4)
		err := common.PropertyWrite(pluginName, appName, property, value)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "get":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		value := common.PropertyGet(pluginName, appName, property)
		if value != "" {
			fmt.Printf("%v\n", value)
		}
	case "get-with-default":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		defaultValue := flag.Arg(4)
		value := common.PropertyGetDefault(pluginName, appName, property, defaultValue)
		if value != "" {
			fmt.Printf("%v\n", value)
		}
	case "get-all":
		appName := flag.Arg(2)
		values, err := common.PropertyGetAll(pluginName, appName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		for key, value := range values {
			fmt.Fprintln(os.Stdout, key, strings.TrimSuffix(value, "\n"))
		}
	case "exists":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		exists := common.PropertyExists(pluginName, appName, property)
		if !exists {
			os.Exit(1)
		}
	case "del":
		appName := flag.Arg(2)
		property := flag.Arg(3)
		err := common.PropertyDelete(pluginName, appName, property)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "setup":
		err := common.PropertySetup(pluginName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "destroy":
		appName := flag.Arg(2)
		err := common.PropertyDestroy(pluginName, appName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

	}
}
