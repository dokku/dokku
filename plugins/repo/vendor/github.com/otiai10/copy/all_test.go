package copy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/otiai10/mint"
)

func TestMain(m *testing.M) {
	setup(m)
	code := m.Run()
	teardown(m)
	os.Exit(code)
}

func setup(m *testing.M) {
	os.MkdirAll("testdata.copy", os.ModePerm)
	os.Symlink("testdata/case01", "testdata/case03/case01")
	os.Chmod("testdata/case07/dir_0500", 0500)
	os.Chmod("testdata/case07/file_0444", 0444)
}

func teardown(m *testing.M) {
	os.RemoveAll("testdata/case03/case01")
	os.RemoveAll("testdata.copy")
}

func TestCopy(t *testing.T) {

	err := Copy("./testdata/case00", "./testdata.copy/case00")
	Expect(t, err).ToBe(nil)
	info, err := os.Stat("./testdata.copy/case00/README.md")
	Expect(t, err).ToBe(nil)
	Expect(t, info.IsDir()).ToBe(false)

	When(t, "specified src doesn't exist", func(t *testing.T) {
		err := Copy("NOT/EXISTING/SOURCE/PATH", "anywhere")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "specified src is just a file", func(t *testing.T) {
		err := Copy("testdata/case01/README.md", "testdata.copy/case01/README.md")
		Expect(t, err).ToBe(nil)
	})

	When(t, "too long name is given", func(t *testing.T) {
		dest := "foobar"
		for i := 0; i < 8; i++ {
			dest = dest + dest
		}
		err := Copy("testdata/case00", filepath.Join("testdata/case00", dest))
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})

	When(t, "try to create not permitted location", func(t *testing.T) {
		err := Copy("testdata/case00", "/case00")
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})

	When(t, "try to create a directory on existing file name", func(t *testing.T) {
		err := Copy("testdata/case02", "testdata.copy/case00/README.md")
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})

	When(t, "source directory includes symbolic link", func(t *testing.T) {
		err := Copy("testdata/case03", "testdata.copy/case03")
		Expect(t, err).ToBe(nil)
		info, err := os.Lstat("testdata.copy/case03/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(0)
	})

	When(t, "symlink with Opt.OnSymlink provided", func(t *testing.T) {
		opt := Options{OnSymlink: func(string) SymlinkAction { return Deep }}
		err := Copy("testdata/case03", "testdata.copy/case03.deep", opt)
		Expect(t, err).ToBe(nil)
		info, err := os.Lstat("testdata.copy/case03.deep/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).ToBe(os.FileMode(0))

		opt = Options{OnSymlink: func(string) SymlinkAction { return Shallow }}
		err = Copy("testdata/case03", "testdata.copy/case03.shallow", opt)
		Expect(t, err).ToBe(nil)
		info, err = os.Lstat("testdata.copy/case03.shallow/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))

		opt = Options{OnSymlink: func(string) SymlinkAction { return Skip }}
		err = Copy("testdata/case03", "testdata.copy/case03.skip", opt)
		Expect(t, err).ToBe(nil)
		_, err = os.Stat("testdata.copy/case03.skip/case01")
		Expect(t, os.IsNotExist(err)).ToBe(true)

		err = Copy("testdata/case03", "testdata.copy/case03.default")
		Expect(t, err).ToBe(nil)
		info, err = os.Lstat("testdata.copy/case03.default/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))

		opt = Options{OnSymlink: nil}
		err = Copy("testdata/case03", "testdata.copy/case03.not-specified", opt)
		Expect(t, err).ToBe(nil)
		info, err = os.Lstat("testdata.copy/case03.not-specified/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))
	})

	When(t, "try to copy to an existing path", func(t *testing.T) {
		err := Copy("testdata/case03", "testdata.copy/case03")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy READ-not-allowed source", func(t *testing.T) {
		err := Copy("testdata/doesNotExist", "testdata.copy/doesNotExist")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy a file to existing path", func(t *testing.T) {
		err := Copy("testdata/case04/README.md", "testdata/case04")
		Expect(t, err).Not().ToBe(nil)
		err = Copy("testdata/case04/README.md", "testdata/case04/README.md/foobar")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy a directory that has no write permission and copy file inside along with it", func(t *testing.T) {
		src := "testdata/case05"
		dest := "testdata.copy/case05"
		err := os.Chmod(src, os.FileMode(0555))
		Expect(t, err).ToBe(nil)
		err = Copy(src, dest)
		Expect(t, err).ToBe(nil)
		info, err := os.Lstat(dest)
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0555))
		err = os.Chmod(dest, 0755)
		Expect(t, err).ToBe(nil)
	})

	When(t, "Options.Skip provided", func(t *testing.T) {
		opt := Options{Skip: func(src string) bool {
			switch {
			case strings.HasSuffix(src, "_skip"):
				return true
			case strings.HasSuffix(src, ".gitfake"):
				return true
			default:
				return false
			}
		}}
		err := Copy("testdata/case06", "testdata.copy/case06", opt)
		Expect(t, err).ToBe(nil)
		info, err := os.Stat("./testdata.copy/case06/dir_skip")
		Expect(t, info).ToBe(nil)
		Expect(t, os.IsNotExist(err)).ToBe(true)

		info, err = os.Stat("./testdata.copy/case06/file_skip")
		Expect(t, info).ToBe(nil)
		Expect(t, os.IsNotExist(err)).ToBe(true)

		info, err = os.Stat("./testdata.copy/case06/README.md")
		Expect(t, info).Not().ToBe(nil)
		Expect(t, err).ToBe(nil)

		info, err = os.Stat("./testdata.copy/case06/repo/.gitfake")
		Expect(t, info).ToBe(nil)
		Expect(t, os.IsNotExist(err)).ToBe(true)

		info, err = os.Stat("./testdata.copy/case06/repo/README.md")
		Expect(t, info).Not().ToBe(nil)
		Expect(t, err).ToBe(nil)
	})

	When(t, "Options.AddPermission provided", func(t *testing.T) {

		info, err := os.Stat("testdata/case07/dir_0500")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()).ToBe(os.FileMode(0500) | os.ModeDir)

		info, err = os.Stat("testdata/case07/file_0444")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()).ToBe(os.FileMode(0444))

		opt := Options{AddPermission: 0200}
		err = Copy("testdata/case07", "testdata.copy/case07", opt)
		Expect(t, err).ToBe(nil)

		info, err = os.Stat("testdata.copy/case07/dir_0500")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()).ToBe(os.FileMode(0500|0200) | os.ModeDir)

		info, err = os.Stat("testdata.copy/case07/file_0444")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()).ToBe(os.FileMode(0444 | 0200))
	})

	When(t, "Options.Sync provided", func(t *testing.T) {
		// With Sync option, each file will be flushed to storage on copying.
		// TODO: Since it's a bit hard to simulate real usecases here. This testcase is nonsense.
		opt := Options{Sync: true}
		err = Copy("testdata/case08", "testdata.copy/case08", opt)
		Expect(t, err).ToBe(nil)
	})
}
