### What problem does this PR solve?
Issue Number: close #63851

Problem Summary:
`EXPLAIN ANALYZE` can show stale timing values when coprocessor cache (`copr_cache`) is hit.
When the same query is executed again and served from cache (`copr_cache_hit_ratio: 1.00`), some timing fields can still display values from the previous uncached execution, which makes diagnostics misleading.
`EXPLAIN FOR CONNECTION` is affected in the same way.

### What changed and how does it work?
- Reset statement-scoped execution detail data before each execution.
- Prevent carry-over of stale cop/tikv timing fields when explain output is built from cached coprocessor responses.
- Keep current execution metrics (cache-hit ratio, rpc/fetch durations, processed keys) consistent with the actual run.
- Apply the same correctness to both `EXPLAIN ANALYZE` and `EXPLAIN FOR CONNECTION`.

### Check List

Tests

- [ ] Unit test
- [ ] Integration test
- [x] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.

Manual test (from issue #63851 reproduce case):
1. Prepare data:
   - `create table d (a int primary key);`
   - `insert into d values (1),(2),(3),(4),(5),(6),(7),(8),(9),(10);`
   - `create table t1M (a int auto_increment primary key, b int, c varchar(255));`
   - `insert into t1M (b,c) select d.a, d.a from d, d d2, d d3, d d4, d d5, d d6;`
   - `analyze table t1M;`
2. Run twice:
   - `explain analyze select * from t1M where c = "non-existent";`
   - `explain analyze select * from t1M where c = "non-existent";`
3. Verify the second run (cache hit) does not display stale long timings from the first run.
4. Verify `EXPLAIN FOR CONNECTION <connection id>` also reports non-stale timing details.

Side effects

- [ ] Performance regression: Consumes more CPU
- [ ] Performance regression: Consumes more Memory
- [ ] Breaking backward compatibility

Documentation

- [ ] Affects user behaviors
- [ ] Contains syntax changes
- [ ] Contains variable changes
- [ ] Contains experimental features
- [ ] Changes MySQL compatibility

### Release note

```release-note
Fix stale timing details in EXPLAIN ANALYZE and EXPLAIN FOR CONNECTION when coprocessor cache is hit.
```
