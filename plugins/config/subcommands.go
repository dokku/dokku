package config

// CommandBundle creates a tarball of a .env.d directory
// containing env vars for the app
func CommandBundle(appName string, global bool, merged bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubBundle(appName, merged)
}

// CommandClear unsets all environment variables in use
func CommandClear(appName string, global bool, noRestart bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubClear(appName, noRestart)
}

// CommandExport outputs all env vars (merged or not, global or not)
// in the specified format for consumption by other tools
func CommandExport(appName string, global bool, merged bool, format string) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubExport(appName, merged, format)
}

// CommandGet gets the value for the specified environment variable
func CommandGet(appName string, keys []string, global bool, quoted bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubGet(appName, keys, quoted)
}

// CommandKeys shows the keys set for the specified environment
func CommandKeys(appName string, global bool, merged bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubKeys(appName, merged)
}

// CommandSet sets one or more environment variable pairs
func CommandSet(appName string, pairs []string, global bool, noRestart bool, encoded bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubSet(appName, pairs, noRestart, encoded)
}

// CommandShow pretty-prints the specified environment vaiables
func CommandShow(appName string, global bool, merged bool, shell bool, export bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubShow(appName, merged, shell, export)
}

// CommandUnset unsets one or more keys in a specified environment
func CommandUnset(appName string, keys []string, global bool, noRestart bool) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	return SubUnset(appName, keys, noRestart)
}
