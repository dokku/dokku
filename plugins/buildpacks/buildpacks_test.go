package buildpacks

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func setupTestEnvironment(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "buildpacks-test")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("DOKKU_LIB_ROOT", tmpDir)
	os.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins")
	os.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")

	return tmpDir
}

func teardownTestEnvironment(tmpDir string) {
	os.RemoveAll(tmpDir)
}

func TestValidBuildpackURL(t *testing.T) {
	RegisterTestingT(t)

	url, err := validBuildpackURL("")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Must specify a buildpack"))

	url, err = validBuildpackURL("heroku/nodejs")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/heroku/heroku-buildpack-nodejs.git"))

	url, err = validBuildpackURL("heroku/python")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/heroku/heroku-buildpack-python.git"))

	url, err = validBuildpackURL("heroku-community/apt")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/heroku/heroku-buildpack-apt.git"))

	url, err = validBuildpackURL("https://github.com/heroku/heroku-buildpack-nodejs.git")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/heroku/heroku-buildpack-nodejs.git"))

	url, err = validBuildpackURL("https://github.com/heroku/heroku-buildpack-nodejs")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("https://github.com/heroku/heroku-buildpack-nodejs"))

	url, err = validBuildpackURL("git://github.com/heroku/heroku-buildpack-nodejs.git")
	Expect(err).NotTo(HaveOccurred())
	Expect(url).To(Equal("git://github.com/heroku/heroku-buildpack-nodejs.git"))

	_, err = validBuildpackURL("nodejs")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Invalid buildpack specified"))

	_, err = validBuildpackURL("/nodejs")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("Invalid buildpack specified"))
}

func TestGetBuildpacksEmpty(t *testing.T) {
	RegisterTestingT(t)
	tmpDir := setupTestEnvironment(t)
	defer teardownTestEnvironment(tmpDir)

	buildpacks, err := getBuildpacks("test-app")
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpacks).To(BeEmpty())
}

func TestGetBuildpacksFromProperties(t *testing.T) {
	RegisterTestingT(t)
	tmpDir := setupTestEnvironment(t)
	defer teardownTestEnvironment(tmpDir)

	propDir := filepath.Join(tmpDir, "config", "buildpacks", "test-app")
	Expect(os.MkdirAll(propDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(propDir, "buildpacks"), []byte("https://github.com/heroku/heroku-buildpack-nodejs.git\nhttps://github.com/heroku/heroku-buildpack-python.git\n"), 0644)).To(Succeed())

	buildpacks, err := getBuildpacks("test-app")
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpacks).To(HaveLen(2))
	Expect(buildpacks[0]).To(Equal("https://github.com/heroku/heroku-buildpack-nodejs.git"))
	Expect(buildpacks[1]).To(Equal("https://github.com/heroku/heroku-buildpack-python.git"))
}

func TestGetBuildpacksFromAppJSON(t *testing.T) {
	RegisterTestingT(t)
	tmpDir := setupTestEnvironment(t)
	defer teardownTestEnvironment(tmpDir)

	appJSONDir := filepath.Join(tmpDir, "data", "app-json", "test-app")
	Expect(os.MkdirAll(appJSONDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(appJSONDir, "app.json"), []byte(`{"buildpacks": [{"url": "https://github.com/heroku/heroku-buildpack-nodejs.git"}]}`), 0644)).To(Succeed())

	buildpacks, err := getBuildpacks("test-app")
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpacks).To(HaveLen(1))
	Expect(buildpacks[0]).To(Equal("https://github.com/heroku/heroku-buildpack-nodejs.git"))
}

func TestGetBuildpacksPropertiesOverrideAppJSON(t *testing.T) {
	RegisterTestingT(t)
	tmpDir := setupTestEnvironment(t)
	defer teardownTestEnvironment(tmpDir)

	propDir := filepath.Join(tmpDir, "config", "buildpacks", "test-app")
	Expect(os.MkdirAll(propDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(propDir, "buildpacks"), []byte("https://github.com/heroku/heroku-buildpack-ruby.git\n"), 0644)).To(Succeed())

	appJSONDir := filepath.Join(tmpDir, "data", "app-json", "test-app")
	Expect(os.MkdirAll(appJSONDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(appJSONDir, "app.json"), []byte(`{"buildpacks": [{"url": "https://github.com/heroku/heroku-buildpack-nodejs.git"}]}`), 0644)).To(Succeed())

	buildpacks, err := getBuildpacks("test-app")
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpacks).To(HaveLen(1))
	Expect(buildpacks[0]).To(Equal("https://github.com/heroku/heroku-buildpack-ruby.git"))
}

func TestGetBuildpacksAppJSONEmptyArray(t *testing.T) {
	RegisterTestingT(t)
	tmpDir := setupTestEnvironment(t)
	defer teardownTestEnvironment(tmpDir)

	appJSONDir := filepath.Join(tmpDir, "data", "app-json", "test-app")
	Expect(os.MkdirAll(appJSONDir, 0755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(appJSONDir, "app.json"), []byte(`{"buildpacks": []}`), 0644)).To(Succeed())

	buildpacks, err := getBuildpacks("test-app")
	Expect(err).NotTo(HaveOccurred())
	Expect(buildpacks).To(BeEmpty())
}
