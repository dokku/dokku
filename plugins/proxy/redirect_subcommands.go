package proxy

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// CommandRedirectSet handles `dokku proxy:redirect:set <app> <source>
// <destination> [--code <code>]`. Uses set semantics: omitted flags reset to
// defaults, and the command is a no-op when the resulting state matches the
// current storage.
func CommandRedirectSet(appName string, source string, destination string, code int) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	if source == "" || destination == "" {
		return errors.New("Usage: dokku proxy:redirect:set <app> <source> <destination> [--code <code>]")
	}

	if code == 0 {
		code = DefaultRedirectCode
	}

	redirect := Redirect{
		Source:      NormalizeRedirectHostname(source),
		Destination: NormalizeRedirectHostname(destination),
		Code:        code,
	}
	changed, err := AddRedirect(appName, redirect)
	if err != nil {
		return err
	}

	if redirect.Code != 301 && redirectBackend(appName) == "haproxy" {
		common.LogWarn(fmt.Sprintf(
			"The haproxy proxy backend always responds to redirects with a 301; the %d code for %s will not be used",
			redirect.Code, redirect.Source,
		))
	}

	if !changed {
		return nil
	}
	return applyRedirects(appName)
}

// CommandRedirectRemove handles `dokku proxy:redirect:remove <app> <source>`.
// Removing a source that does not exist is a successful no-op.
func CommandRedirectRemove(appName string, source string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	if source == "" {
		return errors.New("Usage: dokku proxy:redirect:remove <app> <source>")
	}

	changed, err := RemoveRedirect(appName, NormalizeRedirectHostname(source))
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return applyRedirects(appName)
}

// CommandRedirectClear handles `dokku proxy:redirect:clear <app>`. Clearing an
// empty redirect set is a successful no-op.
func CommandRedirectClear(appName string) error {
	if err := common.VerifyAppName(appName); err != nil {
		return err
	}

	changed, err := ClearRedirects(appName)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return applyRedirects(appName)
}

// applyRedirects rebuilds the proxy config and warns when the active backend
// needs further action before the redirect state is live.
func applyRedirects(appName string) error {
	if err := BuildConfig(appName); err != nil {
		return err
	}
	warnRedirectBackend(appName)
	return nil
}

// redirectBackend returns the backend that renders redirects for an app. The
// k3s scheduler renders redirects via its ingress objects regardless of the
// configured proxy type.
func redirectBackend(appName string) string {
	if common.GetAppScheduler(appName) == "k3s" {
		return "k3s"
	}
	return getComputedProxyType(appName)
}

// warnRedirectBackend prints a notice when the redirect state is not applied
// by the proxy-build-config call alone.
func warnRedirectBackend(appName string) {
	backend := redirectBackend(appName)
	switch backend {
	case "nginx":
		return
	case "caddy", "haproxy", "openresty", "traefik":
		common.LogWarn(fmt.Sprintf(
			"Redirects are applied via container labels under the %s proxy backend; run `dokku ps:rebuild %s` to recreate containers and pick up the change",
			backend, appName,
		))
	case "k3s":
		common.LogWarn(fmt.Sprintf(
			"Redirects are applied via ingress resources under the k3s scheduler; deploy the app or run `dokku ps:rebuild %s` to pick up the change",
			appName,
		))
	default:
		common.LogWarn(fmt.Sprintf(
			"Redirects are stored but not rendered by the %s proxy backend",
			backend,
		))
	}
}

// CommandRedirectReport handles `dokku proxy:redirect:report [<app>] [--format
// stdout|json]`. With no app, iterates every dokku app.
func CommandRedirectReport(appName string, format string) error {
	if format != "stdout" && format != "json" {
		return fmt.Errorf("Invalid format specified, supported formats: json, stdout")
	}

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
			if err := reportRedirectsSingleApp(name, format); err != nil {
				return err
			}
		}
		return nil
	}

	if err := common.VerifyAppName(appName); err != nil {
		return err
	}
	return reportRedirectsSingleApp(appName, format)
}

type redirectReport struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Code        int    `json:"code"`
}

type redirectAppReport struct {
	App       string           `json:"app"`
	Redirects []redirectReport `json:"redirects"`
}

func reportRedirectsSingleApp(appName string, format string) error {
	redirects, err := GetRedirects(appName)
	if err != nil {
		return err
	}

	if format == "json" {
		report := redirectAppReport{App: appName, Redirects: make([]redirectReport, 0, len(redirects))}
		for _, r := range redirects {
			report.Redirects = append(report.Redirects, redirectReport{
				Source:      r.Source,
				Destination: r.Destination,
				Code:        r.Code,
			})
		}
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	common.LogInfo2Quiet(appName + " proxy redirects information")
	if len(redirects) == 0 {
		common.LogVerbose("Redirects: (none)")
		return nil
	}

	for _, r := range redirects {
		common.LogVerbose(fmt.Sprintf("Redirect %s -> %s (%d)", r.Source, r.Destination, r.Code))
	}
	return nil
}
