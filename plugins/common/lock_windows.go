//go:build windows

package common

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// directoryLockFile names the file inside a locked directory that holds the
// lock on windows, where a directory handle cannot be locked directly. The
// file is left in place after unlocking, as removing it would race other
// callers waiting on the lock.
const directoryLockFile = ".dokku.lock"

// lockDirectory holds an exclusive LockFileEx lock on the lock file inside dir
// for the duration of fn
func lockDirectory(dir string, fn func() error) error {
	file, err := os.OpenFile(filepath.Join(dir, directoryLockFile), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Unable to open %s for locking: %w", dir, err)
	}
	defer file.Close()

	handle := windows.Handle(file.Fd())
	if err := windows.LockFileEx(handle, windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &windows.Overlapped{}); err != nil {
		return fmt.Errorf("Unable to lock %s: %w", dir, err)
	}
	defer windows.UnlockFileEx(handle, 0, 1, 0, &windows.Overlapped{})

	return fn()
}
