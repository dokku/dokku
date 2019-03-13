package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
)

func main() {
	args := flag.NewFlagSet("resource:reserve", flag.ExitOnError)
	processeType := args.String("process-type", "", "process-type: A process type to manage")
	cpu := args.String("cpu", "", "cpu: The amount of cpu to reserve for processes")
	memory := args.String("memory", "", "memory: The amount of memory to reserve for processes")
	memorySwap := args.String("memory-swap", "", "memory-swap: The amount of swap memory to reserve for processes")
	network := args.String("network", "", "network: The amount of network bandwidth to reserve for processes")
	networkIngress := args.String("network-ingress", "", "network-ingress: The amount of ingress network bandwidth to reserve for processes")
	networkEgress := args.String("network-egress", "", "network-egress: The amount of egress network bandwidth to reserve for processes")
	args.Parse(os.Args[2:])

	resources := resource.Resource{
		Cpu:            *cpu,
		Memory:         *memory,
		MemorySwap:     *memorySwap,
		Network:        *network,
		NetworkIngress: *networkIngress,
		NetworkEgress:  *networkEgress,
	}

	err := resource.CommandReserve(args.Args(), *processeType, resources)
	if err != nil {
		common.LogFail(err.Error())
	}
}
