package configenv

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"archive/tar"

	"os"

	common "github.com/dokku/dokku/plugins/common"
	godotenv "github.com/joho/godotenv"
	"github.com/ryanuber/columnize"
)

//ExportFormat types of possible exports
type ExportFormat int

const (
	//Exports format: Sourceable exports
	Exports ExportFormat = iota
	//Envfile format: dotenv file
	Envfile
	//DockerArgs format: --env args for docker
	DockerArgs
	//Shell format: env arguments for shell
	Shell
	//Pretty format: pretty-printed in columns
	Pretty
)

//Env is a representation for global or app environment
type Env struct {
	name     string
	filename string
	env      map[string]string
}

//NewFromString creates an env from the given ENVFILE contents representation
func NewFromString(rep string) (env *Env, err error) {
	envMap, err := godotenv.Unmarshal(rep)
	env = &Env{
		"<unknown>",
		"",
		envMap,
	}
	return
}

//LoadApp loads an environment for the given app
func LoadApp(appName string) (env *Env, err error) {
	appfile, err := getAppFile(appName)
	if err != nil {
		return
	}
	return loadFromFile(appName, appfile)
}

//LoadGlobal loads the global environmen
func LoadGlobal() (*Env, error) {
	return loadFromFile("global", getGlobalFile())
}

//Get an environment variable
func (e *Env) Get(key string) (value string, ok bool) {
	value, ok = e.env[key]
	return
}

//GetDefault an environment variable or a default if it doesnt exist
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

//Len return the number of items in this environment
func (e *Env) Len() int {
	return len(e.env)
}

//Map return the Env as a map
func (e *Env) Map() map[string]string {
	return e.env
}

func (e *Env) String() string {
	return e.EnvfileString()
}

//Merge merges the given environment on top of the reciever
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
	case Exports:
		return e.ExportfileString()
	case Envfile:
		return e.EnvfileString()
	case DockerArgs:
		return e.DockerArgsString()
	case Shell:
		return e.ShellString()
	case Pretty:
		return PrettyPrintEnvEntries("", e.Map())
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
	return e.stringWithPrefixAndSeparator("export ", "\n", true)
}

//DockerArgsString gets the contents of this Env in the form -env=KEY=VALUE --env...
func (e *Env) DockerArgsString() string {
	return e.stringWithPrefixAndSeparator("--env=", " ", true)
}

//ShellString gets the contents of this Env in the form "KEY='value' KEY2='value'"
// for passing the environment in the shell
func (e *Env) ShellString() string {
	return e.stringWithPrefixAndSeparator("", " ", true)
}

//ExportBundle writes a tarfile of the environmnet to the given io.Writer.
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

//SingleQuoteEscape escapes the value as if it were shell-quoted in single quotes
func SingleQuoteEscape(value string) string { // so that 'esc'apped' -> 'esc'\''aped'
	return strings.Replace(value, "'", "'\\''", -1)
}

//PrettyPrintEnvEntries in columns
func PrettyPrintEnvEntries(prefix string, entries map[string]string) (representation string) {
	colConfig := columnize.DefaultConfig()
	colConfig.Prefix = prefix
	colConfig.Delim = "\x00"
	lines := make([]string, 0, len(entries))
	for k, v := range entries {
		lines = append(lines, fmt.Sprintf("%s:\x00%s", k, v))
	}
	return columnize.Format(lines, colConfig)
}

func loadFromFile(name string, filename string) (env *Env, err error) {
	envMap := make(map[string]string)
	if _, err := os.Stat(filename); err == nil {
		envMap, err = godotenv.Read(filename)
	}

	env = &Env{
		name,
		filename,
		envMap,
	}
	return
}

//stringWithPrefixAndSeparator makes a string of the environment
// with the given prefix and separator for each entry
func (e *Env) stringWithPrefixAndSeparator(prefix string, separator string, allowNewlines bool) string {
	keys := e.Keys()
	entries := make([]string, len(keys))
	for i, k := range keys {
		v := SingleQuoteEscape(e.env[k])
		if !allowNewlines {
			v = strings.Replace(v, "\n", "'$'\\n''", -1)
		}
		entries[i] = fmt.Sprintf("%s%s='%s'", prefix, k, v)
	}
	return strings.Join(entries, separator)
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
