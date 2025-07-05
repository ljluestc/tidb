// HandleInfoSchemaUpdate updates information schema.
func (do *Domain) HandleInfoSchemaUpdate(is infoschema.InfoSchema) {
	// ...

	// Update schema metrics
	updateSchemaMetrics(is)
}

// updateSchemaMetrics updates metrics for database count, table count, and schema version
func updateSchemaMetrics(is infoschema.InfoSchema) {
	// Update database count
	dbCount := float64(len(is.AllSchemaNames()))
	metrics.SchemaMetrics.DatabaseCount.Set(dbCount)

	// Update table count
	tableCount := 0
	for _, dbName := range is.AllSchemaNames() {
		tables, err := is.AllTableNames(model.NewCIStr(dbName))
		if err == nil {
			tableCount += len(tables)
		}
	}
	metrics.SchemaMetrics.TableCount.Set(float64(tableCount))

	// Update schema version
	schemaVersion := float64(is.SchemaMetaVersion())
	metrics.SchemaMetrics.SchemaVersion.Set(schemaVersion)
}
