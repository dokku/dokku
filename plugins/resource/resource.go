package resource

type Resource struct {
	Cpu            string `json:"cpu"`
	Memory         string `json:"memory"`
	MemorySwap     string `json:"memory-swap"`
	Network        string `json:"network"`
	NetworkIngress string `json:"network-ingress"`
	NetworkEgress  string `json:"network-egress"`
}

const PluginName = "resource"
