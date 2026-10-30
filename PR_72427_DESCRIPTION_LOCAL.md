executor: fix UnionScan missing txn rows for keep-order scan on unsigned handle

### What problem does this PR solve?

Issue Number: close #72427

Problem Summary:

On a table whose clustered primary key is `BIGINT UNSIGNED`, rows inserted earlier in the same explicit transaction are not returned when the optimizer picks an ordered full table scan:

```
UnionScan
└─TableReader
  └─Selection
    └─TableFullScan  keep order:true
```

Minimal reproduction (based on the report):

```sql
CREATE TABLE csong (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tenant VARCHAR(128) NOT NULL,
  PRIMARY KEY (id) CLUSTERED,
  KEY idx_tenant (tenant)
);
-- some committed rows, then ANALYZE so the optimizer prefers the ordered full scan
BEGIN;
INSERT INTO csong (tenant) VALUES ('csong');                 -- 1 row affected
SELECT * FROM csong WHERE tenant = 'csong';                  -- 1 row   (unordered TableFullScan)
SELECT * FROM csong WHERE tenant = 'csong' ORDER BY id;      -- Empty set  <-- wrong
SELECT * FROM csong WHERE tenant = 'csong' ORDER BY id LIMIT 1; -- Empty set <-- wrong
COMMIT;
SELECT * FROM csong WHERE tenant = 'csong' ORDER BY id;      -- 1 row   (data is now in TiKV)
```

This breaks read-your-writes inside a transaction under both RR and RC. The problem is not in statistics. Statistics only decide which plan runs, so the report saw the bug come and go after different `ANALYZE ... WITH n SAMPLERATE` runs:

| Plan | Result |
| --- | --- |
| `UnionScan -> TableReader -> TableFullScan (keep order:true)` | misses uncommitted rows |
| `UnionScan -> TableReader -> TableFullScan (keep order:false)` | correct |
| `Sort -> UnionScan -> IndexLookUp` | correct |
| any plan after `COMMIT` | correct |

#### Root cause

`UnionScanExec` merges the snapshot rows from its child `TableReaderExecutor` with the rows in the transaction membuffer. It reads the membuffer only within `TableReaderExecutor.kvRanges` (see `buildMemTableReader` in `pkg/executor/mem_reader.go`).

In `TableReaderExecutor.Open` (`pkg/executor/table_reader.go`), when the handle is an unsigned int and `keepOrder` is true, `distsql.SplitRangesAcrossInt64Boundary` splits the ranges into two parts at `MaxInt64`. TiKV orders unsigned handles by their int64 encoding, so the two parts must be requested separately and joined by `tableResultHandler`:

- asc: first part `[0, MaxInt64]`, second part `[MaxInt64+1, MaxUint64]`
- desc: first part `[MaxInt64+1, MaxUint64]`, second part `[0, MaxInt64]`

`Open` then calls `buildRespForGroupedRanges` once per part. Each call **assigns** `e.kvRanges` (`e.kvRanges = sortAndGetKVRangesFromReqs(...)`), so the second call overwrites the first. After `Open`, `e.kvRanges` only covers the second part. For an `ORDER BY id` (asc) full scan, UnionScan only looks at membuffer keys with handle `> MaxInt64`, so every normal auto-increment id written in the transaction is skipped.

The path is only reached when all of these hold:

- the table uses an int handle that is unsigned (`BIGINT UNSIGNED` clustered/int PK);
- the scan keeps order (`ORDER BY` the PK, including `ORDER BY ... LIMIT`);
- the ranges cross `MaxInt64` (a full scan always does);
- the read goes through UnionScan (dirty transaction) on a normal (non-dummy) table.

Without keep order, `SplitRangesAcrossInt64Boundary` returns a single part, so there is only one `buildRespForGroupedRanges` call and nothing is lost. The `dummy` branch in the same function (temporary tables and cached tables) already appends the kv ranges of both parts, in ascending order for desc too. Only the normal TiKV path has the bug.

### What changed and how does it work?

`pkg/executor/table_reader.go`

- Save `e.kvRanges` from the first `buildRespForGroupedRanges` call before building the second part, then combine the two.
- UnionScan expects `kvRanges` in ascending key order: for `desc`, `buildMemTableReader` reverses the slice and then iterates the membuffer backwards. So the kv ranges for the smaller handles must come first: the first part for asc, the second part for desc. This is the same order the `dummy` branch produces.
- The distsql requests and `tableResultHandler` are unchanged. Only the ranges given to UnionScan change, so TiKV reads and plain (non-transactional) `ORDER BY` scans behave the same as before.

`pkg/executor/union_scan_test.go`

- Add `TestIssue72427`. It uses a `BIGINT UNSIGNED` clustered PK with committed rows `1, 3, 18446744073709551613`, then inserts uncommitted rows on both sides of the int64 boundary (`2`, `18446744073709551614`). It checks that the plan contains `UnionScan` and that the following return the merged result in the right order:
  - `ORDER BY id`
  - `ORDER BY id DESC`
  - `ORDER BY id LIMIT 1`

### Check List

Tests <!-- At least one of them must be included. -->

- [x] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.
  > <!-- Or your custom  "No need to test" reasons -->

Before the fix, `TestIssue72427` fails because the uncommitted row `2` (handle `<= MaxInt64`) is missing:

```
Error:      Not equal:
            expected: "[2]\n[3]\n[18446744073709551613]\n[18446744073709551614]\n"
            actual  : "[3]\n[18446744073709551613]\n[18446744073709551614]\n"
Messages:   sql:select id from t use index(primary) where tenant = 'b' order by id
```

After the fix it passes, and so do the existing UnionScan and table reader tests:

```bash
make failpoint-enable
cd pkg/executor
go test -run '^TestIssue72427$' -count=1 -tags=intest,deadlock .
# ok  github.com/pingcap/tidb/pkg/executor
go test -run 'UnionScan|TestIssue72427|TestIssue28073|TestTableReader' -count=1 -tags=intest,deadlock .
# ok  github.com/pingcap/tidb/pkg/executor
# (TestTableReaderRequiredRows, TestDirectReadingWithUnionScan, TestUnionScanForMemBufferReader, TestIssue72427, TestIssue28073)
cd ../..
make failpoint-disable
```

Side effects

- [ ] Performance regression: Consumes more CPU
- [ ] Performance regression: Consumes more Memory
- [ ] Breaking backward compatibility

The change only adds one slice append, and only in the unsigned-handle + keep-order path. The extra memory is the first part's kv ranges, which `e.kvRanges` already held before they were overwritten.

Documentation

- [ ] Affects user behaviors
- [ ] Contains syntax changes
- [ ] Contains variable changes
- [ ] Contains experimental features
- [ ] Changes MySQL compatibility

### Release note

<!-- compatibility change, improvement, bugfix, and new feature need a release note -->

Please refer to [Release Notes Language Style Guide](https://pingcap.github.io/tidb-dev-guide/contribute-to-tidb/release-notes-style-guide.html) to write a quality release note.

```release-note
Fix the issue that a query with `ORDER BY` on a `BIGINT UNSIGNED` clustered primary key might miss rows written earlier in the same transaction when the optimizer chooses an ordered full table scan.
```
