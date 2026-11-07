//go:build !windows

package common

import (
	"os"
	"syscall"
)

// lockFile takes an exclusive flock on file
func lockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
}

// unlockFile releases the flock on file
func unlockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
