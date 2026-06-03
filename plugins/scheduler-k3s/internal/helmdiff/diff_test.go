package helmdiff

import (
	"bytes"
	"strings"
	"testing"
)

const sampleDeployment = `# Source: app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: default
spec:
  replicas: 1
`

const sampleDeploymentChanged = `# Source: app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: default
spec:
  replicas: 3
`

const sampleSecret = `# Source: app/templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: creds
  namespace: default
type: Opaque
data:
  password: c2VjcmV0
`

const sampleSecretChanged = `# Source: app/templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: creds
  namespace: default
type: Opaque
data:
  password: bmV3LXNlY3JldA==
`

func runDiff(t *testing.T, oldManifest, newManifest string, opts *Options) string {
	t.Helper()
	current := Parse([]byte(oldManifest), "default", true)
	proposed := Parse([]byte(newManifest), "default", true)
	var buf bytes.Buffer
	Manifests(current, proposed, opts, &buf)
	return buf.String()
}

func TestManifestsNoChange(t *testing.T) {
	out := runDiff(t, sampleDeployment, sampleDeployment, &Options{OutputContext: 3})
	if out != "" {
		t.Fatalf("expected empty diff for identical manifests, got:\n%s", out)
	}
}

func TestManifestsSingleFieldChange(t *testing.T) {
	out := runDiff(t, sampleDeployment, sampleDeploymentChanged, &Options{OutputContext: 3})
	if !strings.Contains(out, "has changed") {
		t.Fatalf("expected MODIFY header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "replicas: 1") {
		t.Fatalf("expected removed line for replicas: 1, got:\n%s", out)
	}
	if !strings.Contains(out, "replicas: 3") {
		t.Fatalf("expected added line for replicas: 3, got:\n%s", out)
	}
}

func TestManifestsAdditiveWhenOldEmpty(t *testing.T) {
	out := runDiff(t, "", sampleDeployment, &Options{OutputContext: 3})
	if !strings.Contains(out, "has been added") {
		t.Fatalf("expected ADD header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "kind: Deployment") {
		t.Fatalf("expected added line for kind: Deployment, got:\n%s", out)
	}
}

func TestManifestsSecretRedactedByDefault(t *testing.T) {
	out := runDiff(t, sampleSecret, sampleSecretChanged, &Options{OutputContext: -1})
	if strings.Contains(out, "c2VjcmV0") || strings.Contains(out, "bmV3LXNlY3JldA==") {
		t.Fatalf("expected raw base64 secret values to be redacted, got:\n%s", out)
	}
	if !strings.Contains(out, "# (") || !strings.Contains(out, "bytes)") {
		t.Fatalf("expected byte-count placeholders in redacted output, got:\n%s", out)
	}
}

func TestManifestsSecretShownWithShowSecrets(t *testing.T) {
	out := runDiff(t, sampleSecret, sampleSecretChanged, &Options{OutputContext: -1, ShowSecrets: true})
	if !strings.Contains(out, "c2VjcmV0") {
		t.Fatalf("expected old base64 value to appear with ShowSecrets, got:\n%s", out)
	}
	if !strings.Contains(out, "bmV3LXNlY3JldA==") {
		t.Fatalf("expected new base64 value to appear with ShowSecrets, got:\n%s", out)
	}
}

func TestManifestsSecretDecodedWithShowSecretsDecoded(t *testing.T) {
	out := runDiff(t, sampleSecret, sampleSecretChanged, &Options{OutputContext: -1, ShowSecretsDecoded: true})
	if !strings.Contains(out, "password: secret") {
		t.Fatalf("expected old plaintext value with ShowSecretsDecoded, got:\n%s", out)
	}
	if !strings.Contains(out, "password: new-secret") {
		t.Fatalf("expected new plaintext value with ShowSecretsDecoded, got:\n%s", out)
	}
}
