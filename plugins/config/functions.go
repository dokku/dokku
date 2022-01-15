package config

import "fmt"

func export(appName string, global bool, merged bool, format string) error {
	appName, err := getAppNameOrGlobal(appName, global)
	if err != nil {
		return err
	}

	env := getEnvironment(appName, merged)
	exportType := ExportFormatExports
	suffix := "\n"

	exportTypes := map[string]ExportFormat{
		"docker-args":      ExportFormatDockerArgs,
		"docker-args-keys": ExportFormatDockerArgsKeys,
		"envfile":          ExportFormatEnvfile,
		"exports":          ExportFormatExports,
		"json":             ExportFormatJSON,
		"json-list":        ExportFormatJSONList,
		"pack":             ExportFormatPackArgKeys,
		"pretty":           ExportFormatPretty,
		"shell":            ExportFormatShell,
	}

	exportType, ok := exportTypes[format]
	if !ok {
		return fmt.Errorf("Unknown export format: %v", format)
	}

	if exportType == ExportFormatShell {
		suffix = " "
	}

	exported := env.Export(exportType)
	fmt.Print(exported + suffix)
	return nil
}
