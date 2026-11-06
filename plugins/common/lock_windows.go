//go:build windows

package common

import "os"

// lockFile is a no-op: dokku only runs on Linux, and the windows build only
// needs to compile for modules that import this package
func lockFile(file *os.File) error {
	return nil
}

// unlockFile is a no-op, see lockFile
func unlockFile(file *os.File) error {
	return nil
}
