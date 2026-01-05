package executor_test

import (
	"testing"

	"github.com/pingcap/tidb/pkg/testkit"
)

func TestChecksumIssue57732(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("USE test")
	tk.MustExec("create table tb(id int primary key) partition by hash (id) partitions 16")
	tk.MustExec("insert into tb values(1)")
    // For MockStore, each request returns Checksum:1, TotalKvs:1, TotalBytes:1.
    // We expect 16 partitions.
    // If Main Table is scanned: 16 partitions + 1 main = 17 requests.
    // If Main Table is NOT scanned: 16 partitions = 16 requests.
    // Expected Result with Bug: 17
    // Expected Result with Fix: 16
    // We assert the BUG first to confirm reproduction.
	tk.MustQuery("ADMIN CHECKSUM TABLE tb").Check(testkit.Rows("test tb 0 17 17"))
}
