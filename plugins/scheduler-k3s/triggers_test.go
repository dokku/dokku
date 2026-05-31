package scheduler_k3s

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dokku/dokku/plugins/common"
	. "github.com/onsi/gomega"
)

func setupChartMigrationTest(t *testing.T) {
	t.Helper()
	RegisterTestingT(t)
	t.Setenv("PLUGIN_PATH", "/var/lib/dokku/plugins")
	t.Setenv("PLUGIN_ENABLED_PATH", "/var/lib/dokku/plugins/enabled")
	t.Setenv("DOKKU_LIB_ROOT", t.TempDir())
	t.Setenv("DOKKU_SYSTEM_USER", "root")
	t.Setenv("DOKKU_SYSTEM_GROUP", "root")
	Expect(common.PropertySetup("scheduler-k3s")).To(Succeed())
}

func setupAnnotationsLabelsMigrationTest(t *testing.T, apps ...string) {
	t.Helper()
	setupChartMigrationTest(t)
	dokkuRoot := t.TempDir()
	t.Setenv("DOKKU_ROOT", dokkuRoot)
	for _, appName := range apps {
		Expect(os.MkdirAll(filepath.Join(dokkuRoot, appName), 0755)).To(Succeed())
	}
}

// writeLegacyAnnotationLabelFile seeds a property file directly using
// PropertyListWrite, matching the pre-migration on-disk layout (one
// "key: value" entry per line).
func writeLegacyAnnotationLabelFile(t *testing.T, scope string, property string, entries map[string]string) {
	t.Helper()
	lines := []string{}
	for k, v := range entries {
		lines = append(lines, k+": "+v)
	}
	Expect(common.PropertyListWrite("scheduler-k3s", scope, property, lines)).To(Succeed())
}

func TestMigrateChartPropertiesToMapFormat(t *testing.T) {
	setupChartMigrationTest(t)

	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.installCRDs", "false")).To(Succeed())
	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.version", "1.13.3")).To(Succeed())
	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.traefik.replicas", "2")).To(Succeed())

	Expect(migrateChartPropertiesToMapFormat()).To(Succeed())

	certManager, err := common.PropertyMapGet("scheduler-k3s", "--global", "chart-overrides.cert-manager")
	Expect(err).NotTo(HaveOccurred())
	Expect(certManager).To(Equal(map[string]string{
		"installCRDs": "false",
		"version":     "1.13.3",
	}))

	traefik, err := common.PropertyMapGet("scheduler-k3s", "--global", "chart-overrides.traefik")
	Expect(err).NotTo(HaveOccurred())
	Expect(traefik).To(Equal(map[string]string{"replicas": "2"}))

	Expect(common.PropertyExists("scheduler-k3s", "--global", "chart.cert-manager.installCRDs")).To(BeFalse())
	Expect(common.PropertyExists("scheduler-k3s", "--global", "chart.cert-manager.version")).To(BeFalse())
	Expect(common.PropertyExists("scheduler-k3s", "--global", "chart.traefik.replicas")).To(BeFalse())
}

func TestMigrateChartPropertiesToMapFormatIsIdempotent(t *testing.T) {
	setupChartMigrationTest(t)

	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.version", "1.13.3")).To(Succeed())

	Expect(migrateChartPropertiesToMapFormat()).To(Succeed())
	Expect(migrateChartPropertiesToMapFormat()).To(Succeed())

	got, err := common.PropertyMapGet("scheduler-k3s", "--global", "chart-overrides.cert-manager")
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"version": "1.13.3"}))
}

func TestMigrateChartPropertiesToMapFormatPreservesExistingNonOverlappingKeys(t *testing.T) {
	setupChartMigrationTest(t)

	Expect(common.PropertyMapWrite("scheduler-k3s", "--global", "chart-overrides.cert-manager", map[string]string{
		"installCRDs":                              "true",
		"service.annotations.prometheus.io/scrape": "true",
	})).To(Succeed())
	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.installCRDs", "false")).To(Succeed())
	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.version", "1.13.3")).To(Succeed())

	Expect(migrateChartPropertiesToMapFormat()).To(Succeed())

	got, err := common.PropertyMapGet("scheduler-k3s", "--global", "chart-overrides.cert-manager")
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{
		"installCRDs":                              "false",
		"service.annotations.prometheus.io/scrape": "true",
		"version":                                  "1.13.3",
	}))
}

func TestMigrateChartPropertiesToMapFormatNoLegacyIsNoOp(t *testing.T) {
	setupChartMigrationTest(t)

	Expect(migrateChartPropertiesToMapFormat()).To(Succeed())

	for _, chart := range HelmCharts {
		Expect(common.PropertyExists("scheduler-k3s", "--global", "chart-overrides."+chart.ReleaseName)).To(BeFalse())
	}
}

func TestMigrateAnnotationsLabelsToMapFormat(t *testing.T) {
	setupAnnotationsLabelsMigrationTest(t, "node-js-app")

	writeLegacyAnnotationLabelFile(t, "node-js-app", "web.deployment", map[string]string{
		"prometheus.io/scrape": "true",
		"team":                 "platform",
	})
	writeLegacyAnnotationLabelFile(t, "node-js-app", "labels.global.service", map[string]string{
		"app.kubernetes.io/managed-by": "dokku",
	})
	writeLegacyAnnotationLabelFile(t, "--global", "global.deployment", map[string]string{
		"owner": "ops",
	})

	Expect(migrateAnnotationsLabelsToMapFormat()).To(Succeed())

	appAnnotations, err := common.PropertyMapGet("scheduler-k3s", "node-js-app", "web.deployment")
	Expect(err).NotTo(HaveOccurred())
	Expect(appAnnotations).To(Equal(map[string]string{
		"prometheus.io/scrape": "true",
		"team":                 "platform",
	}))

	appLabels, err := common.PropertyMapGet("scheduler-k3s", "node-js-app", "labels.global.service")
	Expect(err).NotTo(HaveOccurred())
	Expect(appLabels).To(Equal(map[string]string{
		"app.kubernetes.io/managed-by": "dokku",
	}))

	globalAnnotations, err := common.PropertyMapGet("scheduler-k3s", "--global", "global.deployment")
	Expect(err).NotTo(HaveOccurred())
	Expect(globalAnnotations).To(Equal(map[string]string{"owner": "ops"}))

	raw, err := os.ReadFile(filepath.Join(common.MustGetEnv("DOKKU_LIB_ROOT"), "config", "scheduler-k3s", "node-js-app", "web.deployment"))
	Expect(err).NotTo(HaveOccurred())
	var decoded map[string]string
	Expect(json.Unmarshal(raw, &decoded)).To(Succeed())
	Expect(decoded).To(Equal(appAnnotations))
}

func TestMigrateAnnotationsLabelsToMapFormatIsIdempotent(t *testing.T) {
	setupAnnotationsLabelsMigrationTest(t, "node-js-app")

	writeLegacyAnnotationLabelFile(t, "node-js-app", "web.deployment", map[string]string{
		"prometheus.io/scrape": "true",
	})

	Expect(migrateAnnotationsLabelsToMapFormat()).To(Succeed())
	Expect(migrateAnnotationsLabelsToMapFormat()).To(Succeed())

	got, err := common.PropertyMapGet("scheduler-k3s", "node-js-app", "web.deployment")
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(map[string]string{"prometheus.io/scrape": "true"}))
}

func TestMigrateAnnotationsLabelsToMapFormatPreservesAlreadyMigrated(t *testing.T) {
	setupAnnotationsLabelsMigrationTest(t, "node-js-app")

	preserved := map[string]string{
		"prometheus.io/scrape": "true",
		"controller.config":    "line one\nline two\nline three",
	}
	Expect(common.PropertyMapWrite("scheduler-k3s", "node-js-app", "web.deployment", preserved)).To(Succeed())

	Expect(migrateAnnotationsLabelsToMapFormat()).To(Succeed())

	got, err := common.PropertyMapGet("scheduler-k3s", "node-js-app", "web.deployment")
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(preserved))
}

func TestMigrateAnnotationsLabelsToMapFormatSkipsReservedPrefixes(t *testing.T) {
	setupAnnotationsLabelsMigrationTest(t)

	Expect(common.PropertyWrite("scheduler-k3s", "--global", "chart.cert-manager.deployment", "irrelevant")).To(Succeed())
	Expect(common.PropertyWrite("scheduler-k3s", "--global", "node-profile-foo.json", "{}")).To(Succeed())

	Expect(migrateAnnotationsLabelsToMapFormat()).To(Succeed())

	Expect(common.PropertyGet("scheduler-k3s", "--global", "chart.cert-manager.deployment")).To(Equal("irrelevant"))
	Expect(common.PropertyGet("scheduler-k3s", "--global", "node-profile-foo.json")).To(Equal("{}"))
}
