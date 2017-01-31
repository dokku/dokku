package configenv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	common "github.com/dokku/dokku/plugins/common"
)

//Env is a representation for global or app environment
type Env struct {
	name     string
	env      map[string]string
	filename string
}

func (e *Env) String() string {
	return e.EnvfileString()
}

//EnvfileString returns the contents of this Env in ENVFILE format
func (e *Env) EnvfileString() string {
	return e.stringWithEntryPrefix("")
}

//ExportfileString returns the contents of this Env as bash exports
func (e *Env) ExportfileString() string {
	return e.stringWithEntryPrefix("export ")
}

func (e *Env) stringWithEntryPrefix(prefix string) string {
	keys := e.Keys()
	sort.Strings(keys)
	entries := make([]string, len(keys))
	for i, k := range keys {
		entries[i] = fmt.Sprintf("%s%s='%s'", prefix, k, SingleQuoteEscape(e.env[k]))
	}
	return strings.Join(entries, "\n")
}

//SingleQuoteEscape escapes the value as if it were shell-quoted in single quotes
func SingleQuoteEscape(value string) string { // so that 'esc'apped' -> 'esc'\''aped'
	return strings.Replace(value, "'", "'\\''", -1)
}

//NewFromTarget creates an env from the given target. Tareget is either "--global" or an app name
func NewFromTarget(target string) (*Env, error) {
	if target == "--global" {
		return parseEnv("global", getGlobalFile())
	}
	appfile, err := getAppFile(target)
	if err != nil {
		return nil, err
	}
	return parseEnv(target, appfile)
}

//NewFromString creates an env from the given ENVFILE contents representation
func NewFromString(rep string) (*Env, error) {
	return parseEnvFromReader("<unknown>", "", strings.NewReader(rep))
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
func (e *Env) Keys() []string {
	keys := make([]string, 0, len(e.env))
	for k := range e.env {
		keys = append(keys, k)
	}
	return keys
}

//Get an environment variable
func (e *Env) Get(key string) (string, bool) {
	v, ok := e.env[key]
	return v, ok
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

//Map return the Env as a map
func (e *Env) Map() map[string]string {
	return e.env
}

//Write an Env back to the file it was read from as an exportfile
func (e *Env) Write() error {
	if e.filename == "" {
		return errors.New("this Env was created unbound to a file")
	}
	file, err := os.Create(e.filename)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.WriteString(e.ExportfileString())
	return err
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
