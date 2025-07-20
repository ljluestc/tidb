#!/bin/bash

# This script updates all the required dependencies for TiDB

set -e

# Add all missing dependencies for the TiDB project
echo "Downloading required dependencies..."

# Main parser dependencies
go get -u github.com/pingcap/tidb/pkg/parser
go get -u github.com/pingcap/tidb/pkg/parser/ast
go get -u github.com/pingcap/tidb/pkg/parser/charset
go get -u github.com/pingcap/tidb/pkg/parser/format
go get -u github.com/pingcap/tidb/pkg/parser/mysql
go get -u github.com/pingcap/tidb/pkg/parser/opcode
go get -u github.com/pingcap/tidb/pkg/parser/terror
go get -u github.com/pingcap/tidb/pkg/parser/tidb
go get -u github.com/pingcap/tidb/pkg/parser/types
go get -u github.com/pingcap/tidb/pkg/parser/duration

# External dependencies
go get -u github.com/go-ldap/ldap/v3
go get -u sourcegraph.com/sourcegraph/appdash
go get -u sourcegraph.com/sourcegraph/appdash/opentracing

# Run go mod tidy to clean up dependencies
go mod tidy

echo "Dependencies updated successfully!"
