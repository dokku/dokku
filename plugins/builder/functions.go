package builder

import (
	"io/ioutil"
	"os"
	"path"
)

func removeAllContents(basePath string) error {
	dir, err := ioutil.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, d := range dir {
		os.RemoveAll(path.Join([]string{basePath, d.Name()}...))
	}

	return nil
}
