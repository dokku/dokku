package config

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestExportfileRoundtrip(t *testing.T) {
	RegisterTestingT(t)
	env, err := newEnvFromString("HI='ho'\nFoo='Bar'\n\nBaz='BOFF'")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("Baz", "BOFF", "Foo", "Bar", "HI", "ho")))
	Expect(env.String()).To(Equal("Baz=\"BOFF\"\nFoo=\"Bar\"\nHI=\"ho\""))

	env, err = newEnvFromString(`export HI="h\no"`)
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "h\no")))
	Expect(env.String()).To(Equal(`HI="h\no"`))

	env, err = newEnvFromString("HI='ho'\nFOO=\"'\\nBAR=''\"")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho", "FOO", "'\nBAR=''")))
	Expect(env.String()).To(Equal("FOO=\"'\\nBAR=''\"\nHI=\"ho\""))

	env, err = newEnvFromString("FOO='bar' ")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("FOO", "bar")))
	Expect(env.String()).To(Equal(`FOO="bar"`))

}

func TestExportfileErrors(t *testing.T) {
	RegisterTestingT(t)

	_, err := newEnvFromString("F\nOO='bar'") //keys cannot have embedded newlines
	Expect(err).To(HaveOccurred())
}

func TestMerge(t *testing.T) {
	RegisterTestingT(t)
	e, _ := newEnvFromString("FOO='bar'")
	e2, _ := newEnvFromString("BAR='baz'")
	e.Merge(e2)
	Expect(e.Map()).To(Equal(pairs("BAR", "baz", "FOO", "bar")))

	e3, _ := newEnvFromString("FOO='ba \\nz'")
	e.Merge(e3)
	Expect(e.Map()).To(Equal(pairs("BAR", "baz", "FOO", "ba \nz")))
}

func TestExport(t *testing.T) {
	RegisterTestingT(t)
	e, _ := newEnvFromString("BAR='BAZ'\nFOO='b'ar '\nBAZ='a\\nb'")
	Expect(e.Export(ExportFormatEnvfile)).To(Equal("BAR=\"BAZ\"\nBAZ=\"a\\nb\"\nFOO=\"b'ar \""))
	Expect(e.Export(ExportFormatDockerArgs)).To(Equal("--env=BAR='BAZ' --env=BAZ='a\nb' --env=FOO='b'\\''ar '"))
	Expect(e.Export(ExportFormatDockerArgsKeys)).To(Equal("--env=BAR --env=BAZ --env=FOO"))
	Expect(e.Export(ExportFormatShell)).To(Equal("BAR='BAZ' BAZ='a\nb' FOO='b'\\''ar '"))
	Expect(e.Export(ExportFormatExports)).To(Equal("export BAR='BAZ'\nexport BAZ='a\nb'\nexport FOO='b'\\''ar '"))
	Expect(e.Export(ExportFormatPretty)).To(Equal("BAR:  BAZ\nBAZ:  a\nb\nFOO:  b'ar"))
}

func TestGet(t *testing.T) {
	RegisterTestingT(t)
	e, err := newEnvFromString("BAR='BAZ'\nFOO='ba\\nr '\nGO='1'\nNOGO='0'")
	Expect(err).To(Succeed())

	v, ok := e.Get("GO")
	Expect(ok).To(Equal(true))
	Expect(v).To(Equal("1"))

	v = e.GetDefault("GO", "default")
	Expect(v).To(Equal("1"))

	v = e.GetDefault("dne", "default")
	Expect(v).To(Equal("default"))

	b := e.GetBoolDefault("dne", true)
	Expect(b).To(Equal(true))

	b = e.GetBoolDefault("dne", false)
	Expect(b).To(Equal(false))

	b = e.GetBoolDefault("GO", false)
	Expect(b).To(Equal(true))

	b = e.GetBoolDefault("NOGO", true)
	Expect(b).To(Equal(false))

	b = e.GetBoolDefault("BAR", false)
	Expect(b).To(Equal(true)) //anything but "0" is true
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
