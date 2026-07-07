package certs

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

var dnsPrefixRegex = regexp.MustCompile(`[[:space:]]*DNS:`)

// ReportSingleApp is an internal function that displays the ssl report for one or more apps
func ReportSingleApp(appName string, format string, infoFlag string) error {
	flags := map[string]common.ReportFunc{}
	if appName != "--global" {
		if err := common.VerifyAppName(appName); err != nil {
			return err
		}
		flags = map[string]common.ReportFunc{
			"--ssl-dir":        reportSSLDir,
			"--ssl-enabled":    reportSSLEnabled,
			"--ssl-hostnames":  reportSSLHostnames,
			"--ssl-expires-at": reportSSLExpiresAt,
			"--ssl-issuer":     reportSSLIssuer,
			"--ssl-starts-at":  reportSSLStartsAt,
			"--ssl-subject":    reportSSLSubject,
			"--ssl-verified":   reportSSLVerified,
		}
	}

	flagKeys := []string{}
	for flagKey := range flags {
		flagKeys = append(flagKeys, flagKey)
	}

	infoFlags := common.CollectReport(appName, infoFlag, flags)
	return common.ReportSingleApp(common.ReportSingleAppInput{
		ReportType:              "ssl",
		AppName:                 appName,
		InfoFlag:                infoFlag,
		InfoFlags:               infoFlags,
		InfoFlagKeys:            flagKeys,
		Format:                  format,
		TrimPrefix:              true,
		UppercaseFirstCharacter: true,
		EmitLegacyPrefix:        false,
	})
}

func certTLSPath(appName string) string {
	return filepath.Join(common.MustGetEnv("DOKKU_ROOT"), appName, "tls")
}

func isSSLEnabled(appName string) bool {
	tlsPath := certTLSPath(appName)
	if _, err := os.Stat(filepath.Join(tlsPath, "server.crt")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(tlsPath, "server.key")); err != nil {
		return false
	}

	return true
}

func opensslCertText(appName string) string {
	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "openssl",
		Args:    []string{"x509", "-in", filepath.Join(certTLSPath(appName), "server.crt"), "-noout", "-text"},
	})
	if err != nil {
		return ""
	}

	return result.Stdout
}

func reportSSLDir(appName string) string {
	return certTLSPath(appName)
}

func reportSSLEnabled(appName string) string {
	if isSSLEnabled(appName) {
		return "true"
	}

	return "false"
}

func reportSSLExpiresAt(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	for _, line := range strings.Split(opensslCertText(appName), "\n") {
		if strings.Contains(line, "Not After :") {
			parts := strings.Split(line, " : ")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

func reportSSLStartsAt(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	for _, line := range strings.Split(opensslCertText(appName), "\n") {
		if strings.Contains(line, "Not Before:") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	return ""
}

func reportSSLIssuer(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	for _, line := range strings.Split(opensslCertText(appName), "\n") {
		if strings.Contains(line, "Issuer:") {
			line = strings.Replace(line, "Issuer: ", "", 1)
			return strings.TrimLeft(line, " \t")
		}
	}

	return ""
}

func reportSSLSubject(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	result, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "openssl",
		Args:    []string{"x509", "-in", filepath.Join(certTLSPath(appName), "server.crt"), "-noout", "-subject"},
	})
	if err != nil {
		return ""
	}

	subject := strings.Replace(result.StdoutContents(), "subject= ", "", 1)
	subject = strings.TrimPrefix(subject, "/")
	return strings.ReplaceAll(subject, "/", "; ")
}

func reportSSLVerified(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	tlsPath := certTLSPath(appName)
	args := []string{"verify", "-verbose", "-purpose", "sslserver"}
	if _, err := os.Stat(filepath.Join(tlsPath, "server.letsencrypt.crt")); err == nil {
		args = append(args, "-CAfile", filepath.Join(tlsPath, "server.crt"), filepath.Join(tlsPath, "server.letsencrypt.crt"))
	} else {
		args = append(args, filepath.Join(tlsPath, "server.crt"))
	}

	result, _ := common.CallExecCommand(common.ExecCommandInput{
		Command: "openssl",
		Args:    args,
	})

	verifyOutput := ""
	if out := strings.TrimRight(result.Stdout, "\n"); out != "" {
		lines := strings.Split(out, "\n")
		parts := strings.Split(lines[len(lines)-1], ":")
		if len(parts) >= 2 {
			verifyOutput = strings.TrimSpace(parts[1])
		}
	}

	if verifyOutput == "OK" {
		return "verified by a certificate authority"
	}

	return "self signed"
}

func reportSSLHostnames(appName string) string {
	if !isSSLEnabled(appName) {
		return ""
	}

	hostnameSet := map[string]bool{}

	subjectResult, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "openssl",
		Args:    []string{"x509", "-in", filepath.Join(certTLSPath(appName), "server.crt"), "-noout", "-subject"},
	})
	if err == nil {
		for _, part := range strings.Split(subjectResult.StdoutContents(), "/") {
			if strings.Contains(part, "CN=") && len(part) > 3 {
				hostnameSet[part[3:]] = true
			}
		}
	}

	textLines := strings.Split(opensslCertText(appName), "\n")
	for i, line := range textLines {
		if strings.Contains(line, "509v3 Subject Alternative Name:") && i+1 < len(textLines) {
			sanLine := dnsPrefixRegex.ReplaceAllString(textLines[i+1], "")
			for _, name := range strings.Split(sanLine, ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					hostnameSet[name] = true
				}
			}
		}
	}

	hostnames := []string{}
	for name := range hostnameSet {
		hostnames = append(hostnames, name)
	}
	sort.Strings(hostnames)

	return strings.Join(hostnames, " ")
}
