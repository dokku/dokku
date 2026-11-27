package config

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/joho/godotenv"
	"github.com/ryanuber/columnize"
)

// ExportFormat types of possible exports
type ExportFormat int

const (
	//ExportFormatExports format: Sourceable exports
	ExportFormatExports ExportFormat = iota
	//ExportFormatEnvfile format: dotenv file
	ExportFormatEnvfile
	//ExportFormatDockerArgs format: --env KEY=VALUE args for docker
	ExportFormatDockerArgs
	//ExportFormatDockerArgsKeys format: --env=KEY args for docker
	ExportFormatDockerArgsKeys
	//ExportFormatShell format: env arguments for shell
	ExportFormatShell
	//ExportFormatPretty format: pretty-printed in columns
	ExportFormatPretty
	//ExportFormatJSON format: json key/value output
	ExportFormatJSON
	//ExportFormatJSONList format: json output as a list of objects
	ExportFormatJSONList
	//ExportFormatPackArgKeys format: --env KEY args for pack
	ExportFormatPackArgKeys
)

// Env is a representation for global or app environment
type Env struct {
	name     string
	filename string
	env      map[string]string
}

// newEnvFromString creates an env from the given ENVFILE contents representation
func newEnvFromString(rep string) (env *Env, err error) {
	envMap, err := godotenv.Unmarshal(rep)
	env = &Env{
		name:     "<unknown>",
		filename: "",
		env:      envMap,
	}
	return
}

// LoadAppEnv loads an environment for the given app
func LoadAppEnv(appName string) (env *Env, err error) {
	appfile, err := getAppFile(appName)
	if err != nil {
		return
	}
	return loadFromFile(appName, appfile)
}

// LoadMergedAppEnv loads an app environment merged on top of the app's env
// groups, which are in turn merged on top of the global environment
func LoadMergedAppEnv(appName string) (env *Env, err error) {
	env, err = LoadAppEnv(appName)
	if err != nil {
		return
	}
	global, err := LoadGlobalEnv()
	if err != nil {
		common.LogFailWithError(err)
	}
	groups, err := groupEnvLoader(appName)
	if err != nil {
		return nil, err
	}
	global.Merge(NewEnvFromMap(env.name, groups))
	global.Merge(env)
	global.filename = ""
	global.name = env.name
	return global, nil
}

// groupEnvLoader returns the merged env group layer for an app. It is a
// variable so that tests can run without plugn.
var groupEnvLoader = loadGroupEnv

// loadGroupEnv fetches the merged env group layer for an app via the
// envgroup-app-env trigger. Empty output means the app has no group layer.
func loadGroupEnv(appName string) (map[string]string, error) {
	results, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "envgroup-app-env",
		Args:    []string{appName, "json"},
	})
	if err != nil {
		return nil, fmt.Errorf("Unable to load env groups for %s: %s", appName, err.Error())
	}

	envMap := map[string]string{}
	output := strings.TrimSpace(results.StdoutContents())
	if output == "" {
		return envMap, nil
	}

	if err := json.Unmarshal([]byte(output), &envMap); err != nil {
		return nil, fmt.Errorf("Unable to parse env groups for %s: %s", appName, err.Error())
	}

	return envMap, nil
}

// NewEnvFromMap creates an environment that is not bound to a file from the given map
func NewEnvFromMap(name string, envMap map[string]string) *Env {
	env := &Env{
		name:     name,
		filename: "",
		env:      make(map[string]string, len(envMap)),
	}
	for k, v := range envMap {
		env.env[k] = v
	}
	return env
}

// LoadEnvFile loads the environment stored in the envfile at filename. A
// missing file is treated as an empty environment, and writing the returned
// environment creates it.
func LoadEnvFile(name string, filename string) (*Env, error) {
	return loadFromFile(name, filename)
}

// LoadImportEnv reads an environment to import from filename, or from stdin
// when filename is "-", in the given format. The format defaults to envfile.
func LoadImportEnv(name string, filename string, format string) (*Env, error) {
	switch format {
	case "", "envfile":
		return loadFromFile(name, filename)
	case "json":
		return loadFromFileJSON(name, filename)
	default:
		return nil, fmt.Errorf("Unknown format: %s", format)
	}
}

// ParseExportFormat returns the ExportFormat for its command-line name
func ParseExportFormat(format string) (ExportFormat, error) {
	exportTypes := map[string]ExportFormat{
		"docker-args":      ExportFormatDockerArgs,
		"docker-args-keys": ExportFormatDockerArgsKeys,
		"envfile":          ExportFormatEnvfile,
		"exports":          ExportFormatExports,
		"json":             ExportFormatJSON,
		"json-list":        ExportFormatJSONList,
		"pack-keys":        ExportFormatPackArgKeys,
		"pretty":           ExportFormatPretty,
		"shell":            ExportFormatShell,
	}

	exportType, ok := exportTypes[format]
	if !ok {
		return ExportFormatExports, fmt.Errorf("Unknown export format: %v", format)
	}

	return exportType, nil
}

// LoadGlobalEnv loads the global environment
func LoadGlobalEnv() (*Env, error) {
	return loadFromFile("<global>", getGlobalFile())
}

// Clear clears the environment
func (e *Env) Clear() {
	e.env = make(map[string]string)
}

// Filename returns the full path on disk to the file holding the env vars
func (e *Env) Filename() string {
	return e.filename
}

// Get an environment variable
func (e *Env) Get(key string) (value string, ok bool) {
	value, ok = e.env[key]
	return
}

// GetDefault an environment variable or a default if it doesn't exist
func (e *Env) GetDefault(key string, defaultValue string) string {
	v, ok := e.env[key]
	if !ok {
		return defaultValue
	}
	return v
}

// GetBoolDefault gets the bool value of the given key with the given default
// right now that is evaluated as `value != "0"`
func (e *Env) GetBoolDefault(key string, defaultValue bool) bool {
	v, ok := e.Get(key)
	if !ok {
		return defaultValue
	}
	return v != "0"
}

// Set an environment variable
func (e *Env) Set(key string, value string) {
	e.env[key] = value
}

// Unset an environment variable
func (e *Env) Unset(key string) {
	delete(e.env, key)
}

// Keys gets the keys in this environment
func (e *Env) Keys() (keys []string) {
	keys = make([]string, 0, len(e.env))
	for k := range e.env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

// Len returns the number of items in this environment
func (e *Env) Len() int {
	return len(e.env)
}

// Map returns the Env as a map
func (e *Env) Map() map[string]string {
	return e.env
}

func (e *Env) String() string {
	return e.EnvfileString()
}

// Merge merges the given environment on top of the receiver
func (e *Env) Merge(other *Env) {
	for _, k := range other.Keys() {
		e.Set(k, other.GetDefault(k, ""))
	}
}

// Write an Env back to the file it was read from as an exportfile. The contents
// are written to a temporary file in the same directory and renamed into place,
// so a failed write leaves the previous environment intact rather than
// truncating it. The temporary file is created at 0600, so the values are never
// on disk in a mode other users can read.
func (e *Env) Write() error {
	if e.filename == "" {
		return errors.New("this Env was created unbound to a file")
	}

	contents, err := godotenv.Marshal(e.Map())
	if err != nil {
		return fmt.Errorf("Unable to serialize environment: %s", err.Error())
	}

	target := resolveEnvPath(e.filename)
	file, err := os.CreateTemp(filepath.Dir(target), ".ENV-*")
	if err != nil {
		return fmt.Errorf("Unable to create temporary environment file: %s", err.Error())
	}
	defer os.Remove(file.Name())

	if err := writeEnvFile(file, contents); err != nil {
		return err
	}

	if err := common.SetPermissions(common.SetPermissionInput{
		Filename: file.Name(),
		Mode:     os.FileMode(0600),
	}); err != nil {
		return fmt.Errorf("Unable to set permissions on environment file: %s", err.Error())
	}

	if err := os.Rename(file.Name(), target); err != nil {
		return fmt.Errorf("Unable to move environment file into place: %s", err.Error())
	}

	return nil
}

// resolveEnvPath follows any symlink at filename so the environment is written
// through the symlink rather than replacing it with a regular file, and so the
// temporary file is created on the same filesystem as the real target. A path
// that cannot be resolved - a file that does not exist yet being the common
// case - is returned unchanged.
func resolveEnvPath(filename string) string {
	resolved, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return filename
	}

	return resolved
}

// writeEnvFile writes contents to file, flushes it to disk and closes it
func writeEnvFile(file *os.File, contents string) error {
	if _, err := file.WriteString(contents + "\n"); err != nil {
		file.Close()
		return fmt.Errorf("Unable to write environment file: %s", err.Error())
	}

	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("Unable to flush environment file: %s", err.Error())
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("Unable to close environment file: %s", err.Error())
	}

	return nil
}

// Export the Env in the given format
func (e *Env) Export(format ExportFormat) string {
	switch format {
	case ExportFormatExports:
		return e.ExportfileString()
	case ExportFormatEnvfile:
		return e.EnvfileString()
	case ExportFormatDockerArgs:
		return e.DockerArgsString()
	case ExportFormatDockerArgsKeys:
		return e.DockerArgsKeysString()
	case ExportFormatShell:
		return e.ShellString()
	case ExportFormatPretty:
		return prettyPrintEnvEntries("", e.Map())
	case ExportFormatJSON:
		return e.JSONString()
	case ExportFormatJSONList:
		return e.JSONListString()
	case ExportFormatPackArgKeys:
		return e.PackArgKeysAsString()
	default:
		common.LogFail(fmt.Sprintf("Unknown export format: %v", format))
		return ""
	}
}

// EnvfileString returns the contents of this Env in dotenv format
func (e *Env) EnvfileString() string {
	rep, _ := godotenv.Marshal(e.Map())
	return rep
}

// ExportfileString returns the contents of this Env as bash exports
func (e *Env) ExportfileString() string {
	return e.stringWithPrefixAndSeparator("export ", "\n")
}

// DockerArgsString gets the contents of this Env in the form -env=KEY=VALUE --env...
func (e *Env) DockerArgsString() string {
	return e.stringWithPrefixAndSeparator("--env=", " ")
}

// DockerArgsKeysString gets the contents of this Env in the form -env=KEY --env...
func (e *Env) DockerArgsKeysString() string {
	keys := e.Keys()
	entries := make([]string, len(keys))
	for i, k := range keys {
		entries[i] = fmt.Sprintf("%s%s", "--env=", k)
	}
	return strings.Join(entries, " ")
}

// JSONString returns the contents of this Env as a key/value json object
func (e *Env) JSONString() string {
	data, err := json.Marshal(e.Map())
	if err != nil {
		return "{}"
	}

	return string(data)
}

// JSONListString returns the contents of this Env as a json list of objects containing the name and the value of the env var
func (e *Env) JSONListString() string {
	var list []map[string]string
	for _, key := range e.Keys() {
		value, _ := e.Get(key)
		list = append(list, map[string]string{
			"name":  key,
			"value": value,
		})
	}

	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}

	return string(data)
}

// PackArgKeysAsString gets the contents of this Env in the form -env KEY --env...
func (e *Env) PackArgKeysAsString() string {
	keys := e.Keys()
	entries := make([]string, len(keys))
	for i, k := range keys {
		entries[i] = fmt.Sprintf("%s%s", "--env ", k)
	}
	return strings.Join(entries, " ")
}

// ShellString gets the contents of this Env in the form "KEY='value' KEY2='value'"
// for passing the environment in the shell
func (e *Env) ShellString() string {
	return e.stringWithPrefixAndSeparator("", " ")
}

// ExportBundle writes a tarfile of the environment to the given io.Writer.
// for every environment variable there is a file with the variable's key
// with its content set to the variable's value
func (e *Env) ExportBundle(dest io.Writer) error {
	tarfile := tar.NewWriter(dest)
	defer tarfile.Close()

	for _, k := range e.Keys() {
		val, _ := e.Get(k)
		valbin := []byte(val)

		header := &tar.Header{
			Name: k,
			Mode: 0600,
			Size: int64(len(valbin)),
		}
		tarfile.WriteHeader(header)
		tarfile.Write(valbin)
	}
	return nil
}

// stringWithPrefixAndSeparator makes a string of the environment
// with the given prefix and separator for each entry
func (e *Env) stringWithPrefixAndSeparator(prefix string, separator string) string {
	keys := e.Keys()
	entries := make([]string, len(keys))
	for i, k := range keys {
		v := SingleQuoteEscape(e.env[k])
		entries[i] = fmt.Sprintf("%s%s='%s'", prefix, k, v)
	}
	return strings.Join(entries, separator)
}

// SingleQuoteEscape escapes the value as if it were shell-quoted in single quotes
func SingleQuoteEscape(value string) string { // so that 'esc'aped' -> 'esc'\''aped'
	return strings.Replace(value, "'", "'\\''", -1)
}

// prettyPrintEnvEntries in columns
func prettyPrintEnvEntries(prefix string, entries map[string]string) string {
	colConfig := columnize.DefaultConfig()
	colConfig.Prefix = prefix
	colConfig.Delim = "\x00"

	//some keys may be prefixes of each other so we need to sort them rather than the resulting lines
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s:\x00%s", k, entries[k]))
	}
	return columnize.Format(lines, colConfig)
}

func loadFromFile(name string, filename string) (env *Env, err error) {
	envMap := make(map[string]string)
	dirty := false
	if filename == "-" {
		envMap, err = godotenv.Parse(os.Stdin)
		if err != nil {
			return nil, err
		}
	}
	if _, statErr := os.Stat(filename); statErr == nil {
		contents, readErr := os.ReadFile(filename)
		if readErr != nil {
			return nil, readErr
		}

		parsed, parseErr := godotenv.Unmarshal(string(contents))
		if parseErr != nil {
			common.LogInfo1(fmt.Sprintf("Discarding unparseable entries from config for %s: %s", name, parseErr))
			parsed = parseLineByLine(string(contents))
			dirty = true
		}
		envMap = parsed
	}

	for k := range envMap {
		if err := ValidateKey(k); err != nil {
			common.LogInfo1(fmt.Sprintf("Deleting invalid key %s from config for %s", k, name))
			delete(envMap, k)
			dirty = true
		}
	}
	if dirty {
		if err := godotenv.Write(envMap, filename); err != nil {
			common.LogFail(fmt.Sprintf("Error writing back config for %s after removing invalid keys", name))
		}
	}

	env = &Env{
		name:     name,
		filename: filename,
		env:      envMap,
	}
	return
}

// parseLineByLine parses ENVFILE contents a line at a time, discarding the
// lines the parser rejects, so that a single malformed entry does not make the
// rest of the environment unreadable
func parseLineByLine(contents string) map[string]string {
	envMap := make(map[string]string)
	for _, line := range strings.Split(contents, "\n") {
		lineMap, err := godotenv.Unmarshal(line)
		if err != nil {
			continue
		}
		for key, value := range lineMap {
			envMap[key] = value
		}
	}
	return envMap
}

func loadFromFileJSON(name string, filename string) (env *Env, err error) {
	envMap := make(map[string]string)
	if filename == "-" {
		err = json.NewDecoder(os.Stdin).Decode(&envMap)
		if err != nil {
			return nil, err
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		err = json.NewDecoder(file).Decode(&envMap)
		if err != nil {
			return nil, err
		}
	}

	return &Env{
		name:     name,
		filename: "",
		env:      envMap,
	}, nil
}
func getAppFile(appName string) (string, error) {
	return filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "config", appName, "ENV"), nil
}

func getGlobalFile() string {
	return filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "config", "--global", "ENV")
}
