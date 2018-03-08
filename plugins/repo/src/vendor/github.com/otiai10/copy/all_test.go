package copy

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/otiai10/mint"
)

func TestCopy(t *testing.T) {
	defer os.RemoveAll("testdata/case01")
	err := os.MkdirAll("testdata/case01/bar", os.ModePerm)
	Expect(t, err).ToBe(nil)
	_, err = os.Create("testdata/case01/001.txt")
	Expect(t, err).ToBe(nil)
	_, err = os.Create("testdata/case01/bar/002.txt")
	Expect(t, err).ToBe(nil)
	err = Copy("testdata/case01", "testdata/case01.copy")
	Expect(t, err).ToBe(nil)
	info, err := os.Stat("testdata/case01.copy/bar/002.txt")
	Expect(t, err).ToBe(nil)
	Expect(t, info.IsDir()).ToBe(false)

	When(t, "specified src doesn't exist", func(t *testing.T) {
		err := Copy("not/existing/path", "anywhere")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "specified src is just a file", func(t *testing.T) {
		defer os.RemoveAll("testdata/case01.1")
		os.MkdirAll("testdata/case01.1", os.ModePerm)
		os.Create("testdata/case01.1/001.txt")
		err := Copy("testdata/case01.1/001.txt", "testdata/case01.1/002.txt")
		Expect(t, err).ToBe(nil)
	})

	When(t, "too long name is given", func(t *testing.T) {
		dest := "foobar"
		for i := 0; i < 8; i++ {
			dest = dest + dest
		}
		defer os.RemoveAll("testdata/case02")
		os.MkdirAll("testdata/case02", os.ModePerm)
		os.Create("testdata/case02/001.txt")
		err := Copy("testdata/case02/001.txt", filepath.Join("testdata/case02", dest))
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to create not permitted location", func(t *testing.T) {
		dest := "/001.txt"
		defer os.RemoveAll("testdata/case03")
		os.MkdirAll("testdata/case03", os.ModePerm)
		os.Create("testdata/case03/001.txt")
		err := Copy("testdata/case03/001.txt", dest)
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to create a directory on existing file name", func(t *testing.T) {
		defer os.RemoveAll("testdata/case04")
		os.MkdirAll("testdata/case04", os.ModePerm)
		os.Create("testdata/case04/001.txt")

		os.Create("testdata/case04.copy")

		err := Copy("testdata/case04", "testdata/case04.copy")
		Expect(t, err).Not().ToBe(nil)
	})
}
