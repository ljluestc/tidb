Title: expression: refine negative non-integer constants compared with unsigned integer columns

<!--

Thank you for contributing to TiDB!

PR Title Format:
1. pkg [, pkg2, pkg3]: what's changed
2. *: what's changed

-->

### What problem does this PR solve?
<!--

Please create an issue first to describe the problem.

There MUST be one line starting with "Issue Number:  " and
linking the relevant issues via the "close" or "ref".

For more info, check https://pingcap.github.io/tidb-dev-guide/contribute-to-tidb/contribute-code.html#referring-to-an-issue.

-->

Issue Number: close #11128

Problem Summary:

Part of #11128 was already fixed by #21198. `handleUnsignedCol` in `pkg/util/ranger/points.go` maps an unsigned column compared with a negative constant to `[0,+inf]` or to an empty range. That only helps when the constant reaches the ranger as an integer. A negative **non-integer** constant (decimal or double) still ends up as a full scan plus a filter on a nullable unsigned column:

```sql
create table ti (a int(10) unsigned, key(a));

explain format='brief' select * from ti where a > -100;     -- already fine
IndexReader      root      index:IndexRangeScan
└─IndexRangeScan cop[tikv] table:ti, index:a(a) range:[0,+inf]

explain format='brief' select * from ti where a > -100.1;   -- not fixed
IndexReader       root      index:Selection
└─Selection       cop[tikv] gt(cast(test.ti.a, decimal(10,0) UNSIGNED BINARY), -100.1)
  └─IndexFullScan cop[tikv] table:ti, index:a(a)
```

The same happens for `a < -100.1`, `a >= -100.1`, `a <= -100.1` and for double constants such as `a > -1e2`. A positive decimal is refined correctly (`a > 100.1` gives `range:(100,+inf]`). The plans above are from current `master`.

Root cause: `compareFunctionClass.refineArgs` in `pkg/expression/builtin_compare.go` calls `RefineComparedConstant` with the column's own (unsigned) type. Converting `-100.1` to an unsigned type underflows, so the comparison is marked *exceptional*. For an exceptional comparison, the refined constant is used only when the column is `NOT NULL`, because the rewrite to a constant TRUE/FALSE is wrong for NULL rows. For a nullable column the original decimal constant is kept, the comparison is evaluated as `cast(a as decimal)`, and the ranger cannot build a range from a cast column.

### What changed and how does it work?

In `refineArgs`, for `int column <op> non-int constant` (in either argument order), when:

- the column is an unsigned integer type,
- the constant is negative, and
- `op` is one of `<`, `<=`, `>`, `>=`,

the constant is refined against a **signed BIGINT** type instead of the column type. This uses the same `RefineComparedConstant` ceil/floor rules as positive constants:

| predicate      | refined to     | range from `handleUnsignedCol` |
| -------------- | -------------- | ------------------------------ |
| `a >  -100.1`  | `a >  -101`    | `[0,+inf]`                     |
| `a >= -100.1`  | `a >= -100`    | `[0,+inf]`                     |
| `a <  -100.1`  | `a <  -100`    | empty                          |
| `a <= -100.1`  | `a <= -101`    | empty                          |
| `a >  -1e2`    | `a >  -100`    | `[0,+inf]`                     |

Why this is safe:

- Over the integers, `a > -100.1` and `a > -101` have the same truth value for every value of `a`, and so do the other rows of the table. The column stays nullable and the comparison stays a real comparison, so a NULL `a` still evaluates to NULL. No row's result changes, which is why this works when the column isn't `NOT NULL`, unlike the exceptional rewrite to a constant TRUE/FALSE.
- Comparing an unsigned column with a signed integer constant already works correctly (`types.CompareInt` handles mixed signedness), and the ranger already handles it through `handleUnsignedCol`.
- If ceil/floor of the constant also overflows signed BIGINT (for example `a > -1e30`), behavior is unchanged: the existing exceptional path still applies.
- `=`, `<=>` and `!=` are not changed. Equality with a non-integer is already handled as always false, and `!=` is out of scope.
- Plan cache: this sits behind the existing `allowCmpArgsRefining4PlanCache` / `RemoveMutableConst` guard in `refineArgs`, so parameterized constants are handled the same way as other refinements.

Signed columns and positive constants take the existing path unchanged.

### Check List

Tests <!-- At least one of them must be included. -->

- [x] Unit test
- [x] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.
  > <!-- Or your custom  "No need to test" reasons -->

- `pkg/expression`: unit cases for `refineArgs` covering an unsigned column vs negative decimal/double constants for `<`, `<=`, `>`, `>=` in both argument orders, the `-1e30` overflow case, and that signed columns are unaffected.
- `tests/integrationtest/t/util/ranger.test` (`# TestIssue11128`): `explain` on a nullable unsigned column with a plain index and as the primary key, for the predicates in the table above. Integer scans become `IndexRangeScan`/`TableRangeScan` with `[0,+inf]`, and impossible ranges become `TableDual`. Query results, including rows with NULL `a`, are the same as before the change.

The integration case fails on `master` (it shows `Selection` + `IndexFullScan`) and passes with this change.

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

<!-- compatibility change, improvement, bugfix, and new feature need a release note -->

Please refer to [Release Notes Language Style Guide](https://pingcap.github.io/tidb-dev-guide/contribute-to-tidb/release-notes-style-guide.html) to write a quality release note.

```release-note
Improve the scan range built for an unsigned integer column compared with a negative decimal or floating-point constant, for example `a > -100.1` now scans `[0,+inf]` instead of the full range.
```
