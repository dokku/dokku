package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/dokku/dokku/plugins/common"
)

// CommandRouteSet handles `dokku proxy:route:set <app> <process> <path>
// [--port <port>] [--strip-prefix]`. Uses set semantics: a given invocation
// produces a fully-determined state from its arguments (omitted flags reset
// to defaults), and the command is a no-op when the resulting state is
// byte-identical to current storage.
func CommandRouteSet(appName string, processName string, path string, port int, stripPrefix bool) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if processName == "" || path == "" {
		return errors.New("Usage: dokku proxy:route:set <app> <process> <path> [--port <port>] [--strip-prefix]")
	}

	if port == 0 {
		port = DefaultRoutePort
	}

	route := Route{
		Path:        path,
		Process:     processName,
		Port:        port,
		StripPrefix: stripPrefix,
	}
	changed, err := AddRoute(appName, route)
	if err != nil {
		return err
	}

	if scale := getProcessScale(appName, processName); scale == 0 {
		common.LogWarn(fmt.Sprintf(
			"Process %q is currently scaled to 0; route %s -> %s will resolve to no upstream until the process is scaled up",
			processName, path, processName,
		))
	}

	if !changed {
		return nil
	}
	if err := BuildConfig(appName); err != nil {
		return err
	}
	warnIfRebuildRequired(appName)
	return nil
}

// CommandRouteRemove handles `dokku proxy:route:remove <app> <path>`.
// Removing a path that does not exist is a successful no-op.
func CommandRouteRemove(appName string, path string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	if path == "" {
		return errors.New("Usage: dokku proxy:route:remove <app> <path>")
	}

	changed, err := RemoveRoute(appName, path)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if err := BuildConfig(appName); err != nil {
		return err
	}
	warnIfRebuildRequired(appName)
	return nil
}

// CommandRouteClear handles `dokku proxy:route:clear <app>`. Clearing an
// empty route set is a successful no-op.
func CommandRouteClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	changed, err := ClearRoutes(appName)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if err := BuildConfig(appName); err != nil {
		return err
	}
	warnIfRebuildRequired(appName)
	return nil
}

// warnIfRebuildRequired prints a notice for label-based proxy backends where
// the new route state is only fully applied on the next container start (the
// labels are baked in at create-time).
func warnIfRebuildRequired(appName string) {
	switch getComputedProxyType(appName) {
	case "traefik", "caddy", "openresty":
		common.LogWarn(fmt.Sprintf(
			"Routes are applied via container labels under the %s proxy backend; run `dokku ps:rebuild %s` to recreate containers and pick up the change",
			getComputedProxyType(appName), appName,
		))
	}
}

// CommandRouteReport handles `dokku proxy:route:report [<app>] [--format
// stdout|json]`. With no app, iterates every dokku app.
func CommandRouteReport(appName string, format string) error {
	if appName == "" {
		apps, err := common.DokkuApps()
		if err != nil {
			if errors.Is(err, common.NoAppsExist) {
				common.LogWarn(err.Error())
				return nil
			}
			return err
		}
		for _, name := range apps {
			if err := reportRoutesSingleApp(name, format); err != nil {
				return err
			}
		}
		return nil
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	return reportRoutesSingleApp(appName, format)
}

type routeReport struct {
	Path        string `json:"path"`
	Process     string `json:"process"`
	Port        int    `json:"port"`
	StripPrefix bool   `json:"strip_prefix"`
}

type routeAppReport struct {
	App    string        `json:"app"`
	Routes []routeReport `json:"routes"`
}

func reportRoutesSingleApp(appName string, format string) error {
	routes, err := GetRoutes(appName)
	if err != nil {
		return err
	}

	if format == "json" {
		report := routeAppReport{App: appName, Routes: make([]routeReport, 0, len(routes))}
		for _, r := range routes {
			report.Routes = append(report.Routes, routeReport{
				Path:        r.Path,
				Process:     r.Process,
				Port:        r.Port,
				StripPrefix: r.StripPrefix,
			})
		}
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	common.LogInfo2Quiet(appName + " proxy routes information")
	if len(routes) == 0 {
		common.LogVerbose("Routes: (none)")
		return nil
	}

	keys := make([]string, 0, len(routes))
	for _, r := range routes {
		keys = append(keys, r.Path)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, r := range routes {
			if r.Path != key {
				continue
			}
			strip := ""
			if r.StripPrefix {
				strip = " (strip)"
			}
			common.LogVerbose(fmt.Sprintf("Route %s -> %s:%d%s", r.Path, r.Process, r.Port, strip))
			break
		}
	}
	return nil
}
