// Copyright 2024 PingCAP, Inc.
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
	"strconv"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestChecksum(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("USE test")
	tk.MustExec(`
		CREATE TABLE t (
			c1 INT PRIMARY KEY,
			c2 INT,
			INDEX idx(c2)
		) PARTITION BY RANGE(c1) (
			PARTITION p0 VALUES LESS THAN (10),
			PARTITION p1 VALUES LESS THAN MAXVALUE
		);`)
	tk.MustExec("INSERT INTO t VALUES (1, 11), (2, 12), (15, 13)")

	rows := tk.MustQuery("ADMIN CHECKSUM TABLE t").Rows()
	require.Len(t, rows, 1)

	totalKvs, err := strconv.ParseUint(rows[0][3].(string), 10, 64)
	require.NoError(t, err)
	// 3 table rows + 3 index rows.
	require.Equal(t, uint64(6), totalKvs)

	totalBytes, err := strconv.ParseUint(rows[0][4].(string), 10, 64)
	require.NoError(t, err)
	require.Greater(t, totalBytes, uint64(0))
}

func TestChecksumHashPartitionSingleRow(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("USE test")
	tk.MustExec("CREATE TABLE tb(id INT PRIMARY KEY) PARTITION BY HASH(id) PARTITIONS 16")
	tk.MustExec("INSERT INTO tb VALUES(1)")

	rows := tk.MustQuery("ADMIN CHECKSUM TABLE tb").Rows()
	require.Len(t, rows, 1)

	totalKvs, err := strconv.ParseUint(rows[0][3].(string), 10, 64)
	require.NoError(t, err)
	require.Equal(t, uint64(1), totalKvs)

	totalBytes, err := strconv.ParseUint(rows[0][4].(string), 10, 64)
	require.NoError(t, err)
	require.Greater(t, totalBytes, uint64(0))
}
