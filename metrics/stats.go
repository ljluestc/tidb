// stats.go contains metrics for monitoring

package metrics

import "github.com/prometheus/client_golang/prometheus"

// SchemaMetrics records metrics related to schema information
var SchemaMetrics = struct {
	DatabaseCount prometheus.Gauge
	TableCount    prometheus.Gauge
	SchemaVersion prometheus.Gauge
}{
	DatabaseCount: prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tidb",
			Subsystem: "schema",
			Name:      "database_count",
			Help:      "Number of databases in the TiDB server",
		},
	),
	TableCount: prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tidb",
			Subsystem: "schema",
			Name:      "table_count",
			Help:      "Number of tables in the TiDB server",
		},
	),
	SchemaVersion: prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "tidb",
			Subsystem: "schema",
			Name:      "version",
			Help:      "Current schema version of the TiDB server",
		},
	),
}

func init() {
	// Register schema metrics
	prometheus.MustRegister(SchemaMetrics.DatabaseCount)
	prometheus.MustRegister(SchemaMetrics.TableCount)
	prometheus.MustRegister(SchemaMetrics.SchemaVersion)
}
