package common

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/otiai10/copy"
)

// CatFile cats the contents of a file (if it exists)
func CatFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		LogDebug(fmt.Sprintf("line: '%s'", scanner.Text()))
	}
}

// Copy copies a file/directory from src to dst. If the source is a file, it will also
// convert line endings to unix style
func Copy(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !fi.Mode().IsRegular() {
		return copy.Copy(src, dst)
	}

	// ensure file has the correct line endings
	dos2unixCmd := NewShellCmd(strings.Join([]string{
		"dos2unix",
		"-l",
		"-n",
		src,
		dst,
	}, " "))
	dos2unixCmd.ShowOutput = false
	dos2unixCmd.Execute()

	// ensure file permissions are correct
	b, err := os.ReadFile(dst)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, b, fi.Mode())
	if err != nil {
		return err
	}

	return nil
}

// DirectoryExists returns if a path exists and is a directory
func DirectoryExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// FileToSlice reads in all the lines from a file into a string slice
func FileToSlice(filename string) (lines []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	err = scanner.Err()
	return
}

// FileExists returns if a path exists and is a file
func FileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// IsAbsPath returns 0 if input path is absolute
func IsAbsPath(path string) bool {
	return strings.HasPrefix(path, "/")
}

// ListFilesWithPrefix lists files within a given path that have a given prefix
func ListFilesWithPrefix(path string, prefix string) []string {
	names, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}

	files := []string{}
	for _, f := range names {
		if prefix != "" && !strings.HasPrefix(f.Name(), prefix) {
			continue
		}

		if f.Type().IsRegular() {
			files = append(files, fmt.Sprintf("%s/%s", path, f.Name()))
		}
	}

	return files
}

// ReadFirstLine gets the first line of a file that has contents and returns it
// if there are no contents, an empty string is returned
// will also return an empty string if the file does not exist
func ReadFirstLine(filename string) (text string) {
	if !FileExists(filename) {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if text = strings.TrimSpace(scanner.Text()); text == "" {
			continue
		}
		return
	}
	return
}

// SetPermissions sets the proper owner and filemode for a given file
func SetPermissions(path string, fileMode os.FileMode) error {
	if err := os.Chmod(path, fileMode); err != nil {
		return err
	}

	systemGroup := GetenvWithDefault("DOKKU_SYSTEM_GROUP", "dokku")
	systemUser := GetenvWithDefault("DOKKU_SYSTEM_USER", "dokku")
	if strings.HasPrefix(path, "/etc/sudoers.d/") {
		systemGroup = "root"
		systemUser = "root"
	}

	group, err := user.LookupGroup(systemGroup)
	if err != nil {
		return err
	}
	user, err := user.Lookup(systemUser)
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}
	return os.Chown(path, uid, gid)
}

// TouchFile creates an empty file at the specified path
func TouchFile(filename string) error {
	mode := os.FileMode(0600)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Chmod(mode); err != nil {
		return err
	}

	if err := SetPermissions(filename, mode); err != nil {
		return err
	}
	return nil
}

type WriteSliceToFileInput struct {
	Filename string
	Lines    []string
	Mode     os.FileMode
}

// WriteSliceToFile writes a slice of strings to a file
func WriteSliceToFile(input WriteSliceToFileInput) error {
	file, err := os.OpenFile(input.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, input.Mode)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range input.Lines {
		fmt.Fprintln(w, line)
	}
	if err = w.Flush(); err != nil {
		return err
	}

	if err := file.Chmod(input.Mode); err != nil {
		return err
	}

	if err := SetPermissions(input.Filename, input.Mode); err != nil {
		return err
	}

	return nil
}
