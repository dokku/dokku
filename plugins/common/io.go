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

// SetPermissionsInput is the input struct for SetPermissions
type SetPermissionInput struct {
	Filename  string
	GroupName string
	Mode      os.FileMode
	Username  string
}

// SetPermissions sets the proper owner and filemode for a given file
func SetPermissions(input SetPermissionInput) error {
	if err := os.Chmod(input.Filename, input.Mode); err != nil {
		return err
	}

	if input.GroupName == "" {
		input.GroupName = GetenvWithDefault("DOKKU_SYSTEM_GROUP", "dokku")
	}

	if input.Username == "" {
		input.Username = GetenvWithDefault("DOKKU_SYSTEM_USER", "dokku")
	}

	group, err := user.LookupGroup(input.GroupName)
	if err != nil {
		return err
	}
	user, err := user.Lookup(input.Username)
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
	return os.Chown(input.Filename, uid, gid)
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

	return SetPermissions(SetPermissionInput{
		Filename: filename,
		Mode:     mode,
	})
}

// WriteSliceToFile writes a slice of strings to a file
type WriteSliceToFileInput struct {
	Filename  string
	GroupName string
	Lines     []string
	Mode      os.FileMode
	Username  string
}

// WriteSliceToFile writes a slice of strings to a file
func WriteSliceToFile(input WriteSliceToFileInput) error {
	return WriteBytesToFile(WriteBytesToFileInput{
		Bytes:     []byte(strings.TrimSuffix(strings.Join(input.Lines, "\n"), "\n") + "\n"),
		Filename:  input.Filename,
		GroupName: input.GroupName,
		Mode:      input.Mode,
		Username:  input.Username,
	})
}

// WriteStringToFile writes a string to a file
type WriteStringToFileInput struct {
	Content   string
	Filename  string
	GroupName string
	Mode      os.FileMode
	Username  string
}

// WriteStringToFile writes a string to a file
func WriteStringToFile(input WriteStringToFileInput) error {
	return WriteBytesToFile(WriteBytesToFileInput{
		Bytes:     []byte(input.Content),
		Filename:  input.Filename,
		GroupName: input.GroupName,
		Mode:      input.Mode,
		Username:  input.Username,
	})
}

// WriteBytesToFileInput writes a byte array to a file
type WriteBytesToFileInput struct {
	Bytes     []byte
	Filename  string
	GroupName string
	Mode      os.FileMode
	Username  string
}

// WriteBytesToFile writes a byte array to a file
func WriteBytesToFile(input WriteBytesToFileInput) error {
	file, err := os.OpenFile(input.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, input.Mode)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(input.Bytes); err != nil {
		return err
	}

	if err := file.Chmod(input.Mode); err != nil {
		return err
	}

	permissionsInput := SetPermissionInput{
		Filename: input.Filename,
		Mode:     input.Mode,
	}

	if input.GroupName != "" {
		permissionsInput.GroupName = input.GroupName
	}
	if input.Username != "" {
		permissionsInput.Username = input.Username
	}

	return SetPermissions(permissionsInput)
}
