package sh

import (
	"os"
	"path/filepath"
)

func filetest(name string, modemask os.FileMode) (match bool, err error) {
	fi, err := os.Stat(name)
	if err != nil {
		return
	}
	match = (fi.Mode() & modemask) == modemask
	return
}

func (s *Session) pwd() string {
	dir := string(s.dir)
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return dir
}

func (s *Session) abspath(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(s.pwd(), name)
}

func init() {
	//log.SetFlags(log.Lshortfile | log.LstdFlags)
}

// expression can be dir, file, link
func (s *Session) Test(expression string, argument string) bool {
	var err error
	var fi os.FileInfo
	fi, err = os.Lstat(s.abspath(argument))
	switch expression {
	case "d", "dir":
		return err == nil && fi.IsDir()
	case "f", "file":
		return err == nil && fi.Mode().IsRegular()
	case "x", "executable":
		/*
			fmt.Println(expression, argument)
			if err == nil {
				fmt.Println(fi.Mode())
			}
		*/
		return err == nil && fi.Mode()&os.FileMode(0100) != 0
	case "L", "link":
		return err == nil && fi.Mode()&os.ModeSymlink != 0
	}
	return false
}

// expression can be d,dir, f,file, link
func Test(exp string, arg string) bool {
	s := NewSession()
	return s.Test(exp, arg)
}
