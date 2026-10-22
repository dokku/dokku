package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// LockedDirectoriesEnvVar names the environment variable a caller holding a
// directory lock exports to its child processes. It holds a newline-separated
// list of directories, and LockDirectory runs its callback without locking a
// directory found in that list, so a subprocess spawned while the lock is held
// does not deadlock waiting on its own parent.
const LockedDirectoriesEnvVar = "DOKKU_LOCKED_DIRECTORIES"

// LockDirectory holds an exclusive flock on dir for the duration of fn. The
// directory must already exist. The lock is advisory: it serializes callers
// that go through LockDirectory, and does not block plain reads. Locks are
// taken per open file description, so goroutines in a single process exclude
// each other as well as separate processes do.
//
// A directory listed in DOKKU_LOCKED_DIRECTORIES is treated as already held by
// a parent process, and fn runs without taking the lock again. LockDirectory is
// not reentrant otherwise: fn must not call LockDirectory on the same
// directory, as that would block forever.
func LockDirectory(dir string, fn func() error) error {
	dir = filepath.Clean(dir)
	if isDirectoryLockHeld(dir) {
		return fn()
	}

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

// isDirectoryLockHeld reports whether dir is listed in DOKKU_LOCKED_DIRECTORIES
func isDirectoryLockHeld(dir string) bool {
	for _, held := range strings.Split(os.Getenv(LockedDirectoriesEnvVar), "\n") {
		held = strings.TrimSpace(held)
		if held == "" {
			continue
		}
		if filepath.Clean(held) == dir {
			return true
		}
	}
	return false
}
