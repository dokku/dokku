//go:build !windows

package common

import (
	"fmt"
	"os"
	"syscall"
)

// lockDirectory holds an exclusive flock on dir for the duration of fn
func lockDirectory(dir string, fn func() error) error {
	file, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("Unable to open %s for locking: %w", dir, err)
	}
	defer file.Close()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("Unable to lock %s: %w", dir, err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	return fn()
}
