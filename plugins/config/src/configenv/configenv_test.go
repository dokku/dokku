package configenv

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestExportfileRoundtrip(t *testing.T) {
	RegisterTestingT(t)
	env, err := NewFromString("HI='ho'\nFoo='Bar'\n\nBaz='BOFF'")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("Baz", "BOFF", "Foo", "Bar", "HI", "ho")))
	Expect(env.String()).To(Equal("Baz=\"BOFF\"\nFoo=\"Bar\"\nHI=\"ho\""))

	env, err = NewFromString(`export HI="h\no"`)
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "h\no")))
	Expect(env.String()).To(Equal(`HI="h\no"`))

	env, err = NewFromString("HI='ho'\nFOO=\"'\\nBAR=''\"")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho", "FOO", "'\nBAR=''")))
	Expect(env.String()).To(Equal("FOO=\"'\\nBAR=''\"\nHI=\"ho\""))

	env, err = NewFromString("FOO='bar' ")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("FOO", "bar")))
	Expect(env.String()).To(Equal(`FOO="bar"`))

}

func TestExportfileErrors(t *testing.T) {
	RegisterTestingT(t)

	_, err := NewFromString("F\nOO='bar'") //keys cannot have embedded newlines
	Expect(err).To(HaveOccurred())
}

func TestMerge(t *testing.T) {
	RegisterTestingT(t)
	e, _ := NewFromString("FOO='bar'")
	e2, _ := NewFromString("BAR='baz'")
	e.Merge(e2)
	Expect(e.Map()).To(Equal(pairs("BAR", "baz", "FOO", "bar")))
}

func TestArrayExport(t *testing.T) {
	RegisterTestingT(t)
	e, _ := NewFromString("BAR='BAZ'\nFOO='b'ar '")
	Expect(e.EnvfileString()).To(Equal("BAR=\"BAZ\"\nFOO=\"b'ar \""))
	Expect(e.DockerArgsString()).To(Equal("--env=BAR='BAZ' --env=FOO='b'\\''ar '"))
	Expect(e.ShellString()).To(Equal("BAR='BAZ' FOO='b'\\''ar '"))
}

func pairs(vars ...string) map[string]string {
	res := map[string]string{}
	var i = 0
	for i < len(vars)-1 {
		key := vars[i]
		value := vars[i+1]
		res[key] = value
		i += 2
	}
	return res
}
