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

//ExportFormat types of possible exports
type ExportFormat int

const (
	//ExportFormatExports format: Sourceable exports
	ExportFormatExports ExportFormat = iota
	//ExportFormatEnvfile format: dotenv file
	ExportFormatEnvfile
	//ExportFormatDockerArgs format: --env args for docker
	ExportFormatDockerArgs
	//ExportFormatShell format: env arguments for shell
	ExportFormatShell
	//ExportFormatPretty format: pretty-printed in columns
	ExportFormatPretty
	//ExportFormatJSON format: json key/value output
	ExportFormatJSON
	//ExportFormatJSONList format: json output as a list of objects
	ExportFormatJSONList
)

//Env is a representation for global or app environment
type Env struct {
	name     string
	filename string
	env      map[string]string
}

//newEnvFromString creates an env from the given ENVFILE contents representation
func newEnvFromString(rep string) (env *Env, err error) {
	envMap, err := godotenv.Unmarshal(rep)
	env = &Env{
		name:     "<unknown>",
		filename: "",
		env:      envMap,
	}
	return
}

//LoadAppEnv loads an environment for the given app
func LoadAppEnv(appName string) (env *Env, err error) {
	appfile, err := getAppFile(appName)
	if err != nil {
		return
	}
	return loadFromFile(appName, appfile)
}

//LoadMergedAppEnv loads an app environment merged with the global environment
func LoadMergedAppEnv(appName string) (env *Env, err error) {
	env, err = LoadAppEnv(appName)
	if err != nil {
		return
	}
	global, err := LoadGlobalEnv()
	if err != nil {
		common.LogFail(err.Error())
	}
	global.Merge(env)
	global.filename = ""
	global.name = env.name
	return global, err
}

//LoadGlobalEnv loads the global environment
func LoadGlobalEnv() (*Env, error) {
	return loadFromFile("<global>", getGlobalFile())
}

//Get an environment variable
func (e *Env) Get(key string) (value string, ok bool) {
	value, ok = e.env[key]
	return
}

//GetDefault an environment variable or a default if it doesn't exist
func (e *Env) GetDefault(key string, defaultValue string) string {
	v, ok := e.env[key]
	if !ok {
		return defaultValue
	}
	return v
}

//GetBoolDefault gets the bool value of the given key with the given default
//right now that is evaluated as `value != "0"`
func (e *Env) GetBoolDefault(key string, defaultValue bool) bool {
	v, ok := e.Get(key)
	if !ok {
		return defaultValue
	}
	return v != "0"
}

//Set an environment variable
func (e *Env) Set(key string, value string) {
	e.env[key] = value
}

//Unset an environment variable
func (e *Env) Unset(key string) {
	delete(e.env, key)
}

//Keys gets the keys in this environment
func (e *Env) Keys() (keys []string) {
	keys = make([]string, 0, len(e.env))
	for k := range e.env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

//Len returns the number of items in this environment
func (e *Env) Len() int {
	return len(e.env)
}

//Map returns the Env as a map
func (e *Env) Map() map[string]string {
	return e.env
}

func (e *Env) String() string {
	return e.EnvfileString()
}

//Merge merges the given environment on top of the receiver
func (e *Env) Merge(other *Env) {
	for _, k := range other.Keys() {
		e.Set(k, other.GetDefault(k, ""))
	}
}

//Write an Env back to the file it was read from as an exportfile
func (e *Env) Write() error {
	if e.filename == "" {
		return errors.New("this Env was created unbound to a file")
	}
	return godotenv.Write(e.Map(), e.filename)
}

//Export the Env in the given format
func (e *Env) Export(format ExportFormat) string {
	switch format {
	case ExportFormatExports:
		return e.ExportfileString()
	case ExportFormatEnvfile:
		return e.EnvfileString()
	case ExportFormatDockerArgs:
		return e.DockerArgsString()
	case ExportFormatShell:
		return e.ShellString()
	case ExportFormatPretty:
		return prettyPrintEnvEntries("", e.Map())
	case ExportFormatJSON:
		return e.JSONString()
	case ExportFormatJSONList:
		return e.JSONListString()
	default:
		common.LogFail(fmt.Sprintf("Unknown export format: %v", format))
		return ""
	}
}

//EnvfileString returns the contents of this Env in dotenv format
func (e *Env) EnvfileString() string {
	rep, _ := godotenv.Marshal(e.Map())
	return rep
}

//ExportfileString returns the contents of this Env as bash exports
func (e *Env) ExportfileString() string {
	return e.stringWithPrefixAndSeparator("export ", "\n")
}

//DockerArgsString gets the contents of this Env in the form -env=KEY=VALUE --env...
func (e *Env) DockerArgsString() string {
	return e.stringWithPrefixAndSeparator("--env=", " ")
}

//JSONString returns the contents of this Env as a key/value json object
func (e *Env) JSONString() string {
	data, err := json.Marshal(e.Map())
	if err != nil {
		return "{}"
	}

	return string(data)
}

//JSONListString returns the contents of this Env as a json list of objects containing the name and the value of the env var
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

//ShellString gets the contents of this Env in the form "KEY='value' KEY2='value'"
// for passing the environment in the shell
func (e *Env) ShellString() string {
	return e.stringWithPrefixAndSeparator("", " ")
}

//ExportBundle writes a tarfile of the environment to the given io.Writer.
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

//stringWithPrefixAndSeparator makes a string of the environment
// with the given prefix and separator for each entry
func (e *Env) stringWithPrefixAndSeparator(prefix string, separator string) string {
	keys := e.Keys()
	entries := make([]string, len(keys))
	for i, k := range keys {
		v := singleQuoteEscape(e.env[k])
		entries[i] = fmt.Sprintf("%s%s='%s'", prefix, k, v)
	}
	return strings.Join(entries, separator)
}

//singleQuoteEscape escapes the value as if it were shell-quoted in single quotes
func singleQuoteEscape(value string) string { // so that 'esc'aped' -> 'esc'\''aped'
	return strings.Replace(value, "'", "'\\''", -1)
}

//prettyPrintEnvEntries in columns
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
	if _, err := os.Stat(filename); err == nil {
		envMap, err = godotenv.Read(filename)
	}

	dirty := false
	for k := range envMap {
		if err := validateKey(k); err != nil {
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

func getAppFile(appName string) (string, error) {
	err := common.VerifyAppName(appName)
	if err != nil {
		return "", err
	}
	return filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "ENV"), nil
}

func getGlobalFile() string {
	return filepath.Join(common.MustGetEnv("DOKKU_ROOT"), "ENV")
}
