package executor_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestIssueRepro(t *testing.T) {
	fmt.Println("DEBUG: Starting TestIssueRepro")
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("DROP TABLE IF EXISTS t1")
	tk.MustExec("CREATE TABLE t1 (id tinyint unsigned not null auto_increment primary key)")
	tk.MustExec("INSERT INTO t1 VALUES (255)")
	tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))

	// This insert should fail
	// It tries to allocate the next ID (256), which overflows tinyint unsigned max (255).
	// Depending on strict mode, it might error out or clamp.
	// The user report says "Duplicate entry '255'".
	t.Logf("DEBUG: About to insert NULL")
	_, err := tk.Exec("INSERT INTO t1 VALUES (NULL)")
	t.Logf("DEBUG: Insert error = %v", err)
	require.Error(t, err)
	// require.Contains(t, err.Error(), "Duplicate entry '255' for key")

	// Check last_insert_id(), expect 255
	// The bug is that it returns 256
	// tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))
	// Check last_insert_id(), expect 255
	// The bug is that it returns 256
	// tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))
	// check last_insert_id(), expect 255
	// The bug is that it returns 256
	// tk.MustQuery("SELECT last_insert_id()").Check(testkit.Rows("255"))
	rows := tk.MustQuery("SELECT last_insert_id()").Rows()

	f, _ := os.OpenFile("/tmp/tidb_debug_test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "DEBUG: LastInsertID = %v\n", rows[0][0])
		f.Close()
	}
	require.Equal(t, "255", rows[0][0])
}
