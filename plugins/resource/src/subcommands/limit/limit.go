package main

import (
	"flag"
	"os"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/resource"
)

func main() {
	args := flag.NewFlagSet("resource:limit", flag.ExitOnError)
	processeType := args.String("process-type", "", "process-type: A process type to manage")
	cpu := args.String("cpu", "", "cpu: The amount of cpu to limit processes by")
	memory := args.String("memory", "", "memory: The amount of memory to limit processes by")
	memorySwap := args.String("memory-swap", "", "memory-swap: The amount of swap memory to limit processes by")
	network := args.String("network", "", "network: The amount of network bandwidth to limit processes by")
	networkIngress := args.String("network-ingress", "", "network-ingress: The amount of ingress network bandwidth to limit processes by")
	networkEgress := args.String("network-egress", "", "network-egress: The amount of egress network bandwidth to limit processes by")
	args.Parse(os.Args[2:])

	resources := resource.Resource{
		Cpu:            *cpu,
		Memory:         *memory,
		MemorySwap:     *memorySwap,
		Network:        *network,
		NetworkIngress: *networkIngress,
		NetworkEgress:  *networkEgress,
	}

	err := resource.CommandLimit(args.Args(), *processeType, resources)
	if err != nil {
		common.LogFail(err.Error())
	}
}
