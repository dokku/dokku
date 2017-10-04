package network

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNetworkGetDefaultValue(t *testing.T) {
	RegisterTestingT(t)
	Expect(GetDefaultValue("bind-all-interfaces")).To(Equal("false"))
}
