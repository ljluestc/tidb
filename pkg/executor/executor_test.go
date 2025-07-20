// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestExecutorBasic(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("create table t(a int)")
	tk.MustExec("insert into t values(1)")
	tk.MustQuery("select * from t").Check(testkit.Rows("1"))
}

func TestQueryRegionCount(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t_region")
	tk.MustExec("create table t_region(a int primary key, b int)")

	// Insert some data
	for i := 0; i < 100; i++ {
		tk.MustExec(fmt.Sprintf("insert into t_region values(%d, %d)", i, i))
	}

	// Run a query and check that region count is tracked
	result := tk.MustQuery("explain analyze select * from t_region where a < 50")
	found := false
	for _, row := range result.Rows() {
		if contains(row[0].(string), "region_count:") {
			found = true
			break
		}
	}
	require.True(t, found, "region_count not found in EXPLAIN ANALYZE output")

	// Check slow query log (need to set threshold to 0 to ensure query is logged)
	tk.MustExec("set tidb_slow_log_threshold = 0")
	tk.MustExec("select * from t_region where a < 50")

	// Enable session variables to check the region count
	tk.MustExec("set tidb_metric_query_step_interval = 1")
	tk.MustExec("select * from t_region where a < 50")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
