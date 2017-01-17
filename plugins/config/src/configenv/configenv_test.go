package configenv

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestBasic(t *testing.T) {
	RegisterTestingT(t)
	env, err := NewFromString("HI=ho")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho")))
	Expect(env.String()).To(Equal("HI='ho'"))

	env, err = NewFromString("HI=h\\o\\'")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "h\\o'")))
	Expect(env.String()).To(Equal("HI='h\\o\\''"))

	env, err = NewFromString("HI=ho\nFOO=bar\n\n")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho", "FOO", "bar")))
	Expect(env.String()).To(Equal("FOO='bar'\nHI='ho'"))

	env, err = NewFromString("HI='ho\n'\n\nFOO='bar\\'\nbaz'")
	Expect(err).NotTo(HaveOccurred())
	Expect(env.Map()).To(Equal(pairs("HI", "ho\n", "FOO", "bar'\nbaz")))
	Expect(env.String()).To(Equal("FOO='bar\\'\nbaz'\nHI='ho\n'"))
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
