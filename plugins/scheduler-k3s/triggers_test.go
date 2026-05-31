package scheduler_k3s

import (
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
	Expect(common.PropertySetup("scheduler-k3s")).To(Succeed())
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
