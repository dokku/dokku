package scheduler_k3s

import (
	"encoding/base64"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGetImagePullSecretReleaseName(t *testing.T) {
	got := GetImagePullSecretReleaseName("myapp")
	want := "pull-secret-myapp"
	if got != want {
		t.Fatalf("GetImagePullSecretReleaseName(\"myapp\") = %q, want %q", got, want)
	}
}

func TestGetImagePullSecretName(t *testing.T) {
	got := GetImagePullSecretName("myapp")
	want := "pull-secret-myapp"
	if got != want {
		t.Fatalf("GetImagePullSecretName(\"myapp\") = %q, want %q", got, want)
	}
}

func TestImagePullSecretValuesEncoding(t *testing.T) {
	dockerConfig := []byte(`{"auths":{"https://index.docker.io/v1/":{"auth":"dXNlcjpwYXNz"}}}`)

	values := &ImagePullSecretValues{
		Global: ImagePullSecretGlobalValues{
			AppName:          "myapp",
			Namespace:        "default",
			PullSecretBase64: base64.StdEncoding.EncodeToString(dockerConfig),
		},
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}

	var roundTrip ImagePullSecretValues
	if err := yaml.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("yaml.Unmarshal returned error: %v", err)
	}

	if roundTrip.Global.AppName != "myapp" {
		t.Errorf("AppName = %q, want \"myapp\"", roundTrip.Global.AppName)
	}
	if roundTrip.Global.Namespace != "default" {
		t.Errorf("Namespace = %q, want \"default\"", roundTrip.Global.Namespace)
	}

	decoded, err := base64.StdEncoding.DecodeString(roundTrip.Global.PullSecretBase64)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != string(dockerConfig) {
		t.Errorf("PullSecretBase64 decoded = %q, want %q", string(decoded), string(dockerConfig))
	}
}
