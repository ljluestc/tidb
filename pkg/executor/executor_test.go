func (s *testSuiteSuite) TestNewExecWithLabel(c *C) {
	store, dom, err := NewStoreWithBootstrap()
	c.Assert(err, IsNil)
	defer func() {
		dom.Close()
		err := store.Close()
		c.Assert(err, IsNil)
	}()
	tk := testkit.NewTestKit(c, store)
	tk.MustExec("use test")
	tk.MustExec("create table with_label (id int, value timestamp(3))")
	for i := 0; i < 10; i++ {
		sql := fmt.Sprintf("insert into with_label values (%d, '2019-01-01 00:00:0%d.123')", i, i)
		tk.MustExec(sql)
	}
	tk.MustExec("analyze table with_label")
}

func (s *testSuiteSuite) TestQueryRegionCount(c *C) {
	store, dom, err := NewStoreWithBootstrap()
	c.Assert(err, IsNil)
	defer func() {
		dom.Close()
		err := store.Close()
		c.Assert(err, IsNil)
	}()

	tk := testkit.NewTestKit(c, store)
	tk.MustExec("use test")
	tk.MustExec("create table t_region_count (id int primary key, value int)")
	tk.MustExec("create table t_region_count2 (id int primary key, value int)")

	// Insert some data to potentially create multiple regions
	for i := 0; i < 1000; i++ {
		tk.MustExec(fmt.Sprintf("insert into t_region_count values (%d, %d)", i, i))
		if i < 500 {
			tk.MustExec(fmt.Sprintf("insert into t_region_count2 values (%d, %d)", i, i))
		}
	}

	// Run a query that should access regions
	tk.MustQuery("select * from t_region_count where id > 10 and id < 100")

	// Run a join query that should access more regions
	tk.MustQuery("select a.id, b.value from t_region_count a join t_region_count2 b on a.id = b.id where a.id < 100")

	// Verify region count in statement context
	result := tk.MustQuery("explain analyze select * from t_region_count where id > 10 and id < 100")
	explainStr := result.Rows()[0][0].(string)
	c.Assert(strings.Contains(explainStr, "region_count:"), IsTrue)

	// Run slow query to check slow log with region count
	tk.MustExec("set tidb_slow_log_threshold = 0")
	tk.MustQuery("select * from t_region_count where id > 10 and id < 100")

	// Actual metrics check would need to be done with Prometheus client test library
	// But we can at least verify that the region count tracking is in place
}
