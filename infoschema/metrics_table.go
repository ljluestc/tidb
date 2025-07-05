// MetricTableMap records the metrics table definition
var MetricTableMap = map[string]model.MetricTableDef{
	// Existing metric tables

	// Add new schema metrics tables
	"database_count": {
		PromQL:    "tidb_schema_database_count",
		Labels:    []string{},
		Quantile:  0,
		Comment:   "Number of databases in the TiDB server",
	},
	"table_count": {
		PromQL:    "tidb_schema_table_count",
		Labels:    []string{},
		Quantile:  0,
		Comment:   "Number of tables in the TiDB server",
	},
	"schema_version": {
		PromQL:    "tidb_schema_version",
		Labels:    []string{},
		Quantile:  0,
		Comment:   "Current schema version of the TiDB server",
	},
}
