package resource

// Resource is a collection of resource constraints for apps
type Resource struct {
	CPU            string `json:"cpu"`
	Memory         string `json:"memory"`
	MemorySwap     string `json:"memory-swap"`
	Network        string `json:"network"`
	NetworkIngress string `json:"network-ingress"`
	NetworkEgress  string `json:"network-egress"`
}
