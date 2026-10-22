package common

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCommonIsValidMemoryValue(t *testing.T) {
	RegisterTestingT(t)

	valid := []string{"0", "100", "1.5", "100b", "100B", "512k", "512K", "512kb", "512m", "512M", "512MB", "512Mi", "512MiB", "1g", "1G", "1GB", "1Gi", "1gib", "1t", "1T", "1p", "1P"}
	for _, value := range valid {
		Expect(IsValidMemoryValue(value)).To(BeTrue(), value)
	}

	invalid := []string{"", "-1", "abc", "1x", "1Gm", "1i", "1ib", "1 G", "1.2.3", ".5g", "100G --foo", "1E", "1e9"}
	for _, value := range invalid {
		Expect(IsValidMemoryValue(value)).To(BeFalse(), value)
	}
}

func TestCommonIsUnitlessMemoryValue(t *testing.T) {
	RegisterTestingT(t)

	Expect(IsUnitlessMemoryValue("100")).To(BeTrue())
	Expect(IsUnitlessMemoryValue("1.5")).To(BeTrue())
	Expect(IsUnitlessMemoryValue("100b")).To(BeFalse())
	Expect(IsUnitlessMemoryValue("1G")).To(BeFalse())
	Expect(IsUnitlessMemoryValue("1Gi")).To(BeFalse())
	Expect(IsUnitlessMemoryValue("abc")).To(BeFalse())
}

func TestCommonParseMemoryBytes(t *testing.T) {
	RegisterTestingT(t)

	expected := map[string]int64{
		"0":      0,
		"512":    536870912,
		"512m":   536870912,
		"512M":   536870912,
		"512MB":  536870912,
		"512Mi":  536870912,
		"512MiB": 536870912,
		"1g":     1073741824,
		"1G":     1073741824,
		"1GB":    1073741824,
		"1Gi":    1073741824,
		"1.5g":   1610612736,
		"1024K":  1048576,
		"100b":   100,
		"100B":   100,
		"1t":     1099511627776,
		"1p":     1125899906842624,
	}
	for value, bytes := range expected {
		parsed, err := ParseMemoryBytes(value)
		Expect(err).NotTo(HaveOccurred(), value)
		Expect(parsed).To(Equal(bytes), value)
	}

	for _, value := range []string{"", "-1", "1x", "1Gm", "100G --foo"} {
		_, err := ParseMemoryBytes(value)
		Expect(err).To(HaveOccurred(), value)
	}
}
