### What problem does this PR solve?
Issue Number: close #10472

Problem Summary:
When rebuilding cached `PointGet` / `BatchPointGet` plans for unsigned PK handles, the rebuild path converted handle parameters with `Datum.ToInt64()`. For values beyond `math.MaxInt64` (for example `18446744073709551615`), this may fail rebuild and prevent stable cache reuse.

### What changed and how does it work?
1. In `pkg/planner/core/plan_cache_rebuild.go`, add `datumToIntHandleValue(...)`:
   - unsigned handle type: use `Datum.GetInt64()` (bit-preserving)
   - signed handle type: keep `Datum.ToInt64(...)` behavior
2. Use the real PK unsigned flag (`mysql.HasUnsignedFlag(pkColInfo.GetFlag())`) when deriving `unsignedIntHandle` in point/batch point-get rebuild safety checks.
3. Extend regression coverage in:
   - `tests/integrationtest/t/planner/core/tests/prepare/prepare.test`
   - `tests/integrationtest/r/planner/core/tests/prepare/prepare.result`
   by validating repeated point get and batch point get cache behavior for max `BIGINT UNSIGNED` handle values.

### Check List
Tests <!-- At least one of them must be included. -->
- [x] Unit test
- [x] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test

Executed test commands:
1. `GOFLAGS='-p=1' go test ./pkg/planner/core/casetest/plancache -run TestPlanCacheClone -count=1 -tags=intest,deadlock`
2. `cd tests/integrationtest && GOFLAGS='-p=1' ./run-tests.sh -r planner/core/tests/prepare/prepare`

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
None
```
