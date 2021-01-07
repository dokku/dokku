package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"

	flag "github.com/spf13/pflag"
)

// main entrypoint to all subcommands
func main() {
	parts := strings.Split(os.Args[0], "/")
	subcommand := parts[len(parts)-1]

	var err error
	switch subcommand {
	case "limit":
		args := flag.NewFlagSet("resource:limit", flag.ExitOnError)
		processType := args.String("process-type", "", "process-type: A process type to manage")
		cpu := args.String("cpu", "", "cpu: The amount of cpu to limit processes by")
		memory := args.String("memory", "", "memory: The amount of memory to limit processes by")
		memorySwap := args.String("memory-swap", "", "memory-swap: The amount of swap memory to limit processes by")
		network := args.String("network", "", "network: The amount of network bandwidth to limit processes by")
		networkIngress := args.String("network-ingress", "", "network-ingress: The amount of ingress network bandwidth to limit processes by")
		networkEgress := args.String("network-egress", "", "network-egress: The amount of egress network bandwidth to limit processes by")
		nvidiaGpu := args.String("nvidia-gpu", "", "nvidia-gpu: The number of Nvidia GPUs to limit a process to")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)

		resources := resource.Resource{
			CPU:            *cpu,
			Memory:         *memory,
			MemorySwap:     *memorySwap,
			Network:        *network,
			NetworkIngress: *networkIngress,
			NetworkEgress:  *networkEgress,
			NvidiaGPU:      *nvidiaGpu,
		}

		err = resource.CommandLimit(appName, *processType, resources)
	case "limit-clear":
		args := flag.NewFlagSet("resource:limit-clear", flag.ExitOnError)
		processType := args.String("process-type", "", "process-type: A process type to clear")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = resource.CommandLimitClear(appName, *processType)
	case "report":
		args := flag.NewFlagSet("resource:report", flag.ExitOnError)
		osArgs, infoFlag, flagErr := common.ParseReportArgs("resource", os.Args[2:])
		if flagErr == nil {
			args.Parse(osArgs)
			appName := args.Arg(0)
			err = resource.CommandReport(appName, infoFlag)
		}
	case "reserve":
		args := flag.NewFlagSet("resource:reserve", flag.ExitOnError)
		processType := args.String("process-type", "", "process-type: A process type to manage")
		cpu := args.String("cpu", "", "cpu: The amount of cpu to reserve for processes")
		memory := args.String("memory", "", "memory: The amount of memory to reserve for processes")
		memorySwap := args.String("memory-swap", "", "memory-swap: The amount of swap memory to reserve for processes")
		network := args.String("network", "", "network: The amount of network bandwidth to reserve for processes")
		networkIngress := args.String("network-ingress", "", "network-ingress: The amount of ingress network bandwidth to reserve for processes")
		networkEgress := args.String("network-egress", "", "network-egress: The amount of egress network bandwidth to reserve for processes")
		nvidiaGpu := args.String("nvidia-gpu", "", "nvidia-gpu: The number of Nvidia GPUs to resource for a process")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)

		resources := resource.Resource{
			CPU:            *cpu,
			Memory:         *memory,
			MemorySwap:     *memorySwap,
			Network:        *network,
			NetworkIngress: *networkIngress,
			NetworkEgress:  *networkEgress,
			NvidiaGPU:      *nvidiaGpu,
		}

		err = resource.CommandReserve(appName, *processType, resources)
	case "reserve-clear":
		args := flag.NewFlagSet("resource:reserve-clear", flag.ExitOnError)
		processType := args.String("process-type", "", "process-type: A process type to clear")
		args.Parse(os.Args[2:])
		appName := args.Arg(0)
		err = resource.CommandReserveClear(appName, *processType)
	default:
		common.LogFail(fmt.Sprintf("Invalid plugin subcommand call: %s", subcommand))
	}

	if err != nil {
		common.LogFail(err.Error())
	}
}
