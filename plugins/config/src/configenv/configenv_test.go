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
	Expect(env.String()).To(Equal("Baz='BOFF'\nFoo='Bar'\nHI='ho'"))

	env, err = NewFromString("\n export HI='h\\\no\\'\n")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "h\\\no\\")))
	Expect(env.String()).To(Equal("HI='h\\\no\\'"))

	env, err = NewFromString("HI='ho'\nFOO=''\\''\nBAR='\\'''\\'''")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho", "FOO", "'\nBAR=''")))
	Expect(env.String()).To(Equal("FOO=''\\''\nBAR='\\'''\\'''\nHI='ho'"))

	env, err = NewFromString("HI='ho\n'\n\nFOO='=bar\"\n'\\''\nbaz'")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho\n", "FOO", "=bar\"\n'\nbaz")))
	Expect(env.String()).To(Equal("FOO='=bar\"\n'\\''\nbaz'\nHI='ho\n'"))
}

func TestExportfileErrors(t *testing.T) {
	RegisterTestingT(t)
	_, err := NewFromString("FOO=bar") //values must be quoted
	Expect(err).To(HaveOccurred())

	_, err = NewFromString("FOO='bar\\''") //single quotes are not escaped this way
	Expect(err).To(HaveOccurred())

	_, err = NewFromString("F\nOO='bar'") //keys cannot have embedded newlines
	Expect(err).To(HaveOccurred())

	_, err = NewFromString("FOO='bar' ") //no trailing content
	Expect(err).To(HaveOccurred())
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
