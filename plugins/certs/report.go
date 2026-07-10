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
		Args:    []string{"x509", "-in", filepath.Join(certTLSPath(appName), "server.crt"), "-noout", "-subject", "-nameopt", "compat"},
	})
	if err != nil {
		return ""
	}

	return formatSSLSubject(result.StdoutContents())
}

// formatSSLSubject normalizes an openssl "-subject -nameopt compat" line into a
// "; "-joined RDN string. The compat nameopt reproduces the legacy "/"-delimited,
// order-preserving subject form on every OpenSSL/LibreSSL version.
func formatSSLSubject(out string) string {
	out = strings.TrimPrefix(strings.TrimSpace(out), "subject=")
	out = strings.TrimPrefix(strings.TrimSpace(out), "/")
	return strings.ReplaceAll(out, "/", "; ")
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

	subject := ""
	subjectResult, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "openssl",
		Args:    []string{"x509", "-in", filepath.Join(certTLSPath(appName), "server.crt"), "-noout", "-subject", "-nameopt", "RFC2253"},
	})
	if err == nil {
		subject = subjectResult.StdoutContents()
	}

	return strings.Join(sslHostnames(subject, opensslCertText(appName)), " ")
}

// subjectCommonName extracts the CN value from an openssl "-subject" line. It is
// tolerant of the "CN=value" (RFC2253) and "CN = value" (OpenSSL 3.x default)
// renderings, the legacy "/"-delimited compat form, and a leading "subject="
// prefix. The CN value of a certificate is a hostname, so splitting on "/" as
// well as "," never truncates it.
func subjectCommonName(subject string) string {
	subject = strings.TrimPrefix(strings.TrimSpace(subject), "subject=")
	for _, rdn := range strings.FieldsFunc(subject, func(r rune) bool { return r == ',' || r == '/' }) {
		key, value, found := strings.Cut(rdn, "=")
		if found && strings.TrimSpace(key) == "CN" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// sslHostnames returns the sorted, de-duplicated set of hostnames a certificate
// covers, merging the subject Common Name (from an RFC2253 "-subject" line) with
// every Subject Alternative Name DNS entry (from the "-text" rendering).
func sslHostnames(subject string, certText string) []string {
	hostnameSet := map[string]bool{}
	if cn := subjectCommonName(subject); cn != "" {
		hostnameSet[cn] = true
	}

	textLines := strings.Split(certText, "\n")
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

	return hostnames
}
