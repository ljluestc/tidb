package executor_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestLastInsertIDOverflow(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("DROP TABLE IF EXISTS t1")
	tk.MustExec("CREATE TABLE t1 (id tinyint unsigned not null auto_increment primary key)")
	
	// Insert 254 explicitly
	tk.MustExec("INSERT INTO t1 VALUES (254)")
	
	// Insert 255 via auto-increment
	tk.MustExec("INSERT INTO t1 VALUES (NULL)")
	tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))

	// This insert should fail
	// It tries to allocate the next ID (256), which overflows tinyint unsigned max (255).
	_, err := tk.Exec("INSERT INTO t1 VALUES (NULL)")
	require.Error(t, err)
	// require.Contains(t, err.Error(), "Duplicate entry '255' for key")

	// Check last_insert_id(), expect 255. PREVIOUSLY it returned 256.
	tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))
}
