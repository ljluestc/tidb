package infoschema_test

import (
	"testing"
	// existing imports
	"github.com/pingcap/tidb/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// TestSchemaMetrics tests that schema metrics are properly updated
func TestSchemaMetrics(t *testing.T) {
	store, dom := testkit.CreateMockStoreAndDomain(t)
	defer func() {
		dom.Close()
		store.Close()
	}()

	tk := testkit.NewTestKit(t, store)

	// Get initial metric values
	initialDBCount := getGaugeValue(t, metrics.SchemaMetrics.DatabaseCount)
	initialTableCount := getGaugeValue(t, metrics.SchemaMetrics.TableCount)
	initialSchemaVersion := getGaugeValue(t, metrics.SchemaMetrics.SchemaVersion)

	// Create a new database
	tk.MustExec("CREATE DATABASE test_metrics")

	// Check database count increased
	newDBCount := getGaugeValue(t, metrics.SchemaMetrics.DatabaseCount)
	require.Equal(t, initialDBCount+1, newDBCount, "Database count metric should increase after creating a database")

	// Create a new table
	tk.MustExec("USE test_metrics")
	tk.MustExec("CREATE TABLE test_table (id INT)")

	// Check table count increased
	newTableCount := getGaugeValue(t, metrics.SchemaMetrics.TableCount)
	require.Equal(t, initialTableCount+1, newTableCount, "Table count metric should increase after creating a table")

	// Check schema version increased
	newSchemaVersion := getGaugeValue(t, metrics.SchemaMetrics.SchemaVersion)
	require.Greater(t, newSchemaVersion, initialSchemaVersion, "Schema version should increase after schema changes")

	// Drop database
	tk.MustExec("DROP DATABASE test_metrics")

	// Verify counts decreased
	finalDBCount := getGaugeValue(t, metrics.SchemaMetrics.DatabaseCount)
	finalTableCount := getGaugeValue(t, metrics.SchemaMetrics.TableCount)

	require.Equal(t, initialDBCount, finalDBCount, "Database count should return to initial value")
	require.Equal(t, initialTableCount, finalTableCount, "Table count should return to initial value")
}

// getGaugeValue extracts the value from a Prometheus gauge
func getGaugeValue(t *testing.T, gauge prometheus.Gauge) float64 {
	metric := &dto.Metric{}
	err := gauge.Write(metric)
	require.NoError(t, err)
	return metric.GetGauge().GetValue()
}
