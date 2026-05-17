# fix(transaction): prevent write skew cycles with SELECT ... FOR UPDATE on non-existent rows

Closes #10444

## Problem

TiDB's `SELECT ... FOR UPDATE` mechanism is unable to prevent write skew anomalies when transactions read non-existent rows. This violates serializability guarantees when using pessimistic locking.

### Issue Details

Jepsen testing discovered anti-dependency cycles and write skew anomalies in TiDB 2.1.7, 2.1.8, 3.0.0-beta.1, and 3.0.0-beta1.40, even when all reads use `SELECT ... FOR UPDATE`:

**Example 1: Anti-dependency cycle**
```
T1 : ... r(1067, nil), append(1066, 1)
T2 : ... r(1066, nil), append(1067, 1)
```

- T1 < T2: because T1 observed initial nil state of 1067, which T2 later wrote
- T2 < T1: because T2 observed initial nil state of 1066, which T1 later wrote
- **Contradiction:** Anti-dependency cycle allowed under snapshot isolation but forbidden with serializable isolation

**Example 2: Write skew**
```
T1 = [r(3, nil), r(4, nil), append(4, 2)]
T2 = [r(3, nil), r(4, nil), append(3, 1)]
```

- T1 < T2: T1 observed nil for key 3, which T2 created
- T2 < T1: T2 observed nil for key 4, which T1 created
- **Contradiction:** Write skew cycle impossible under serializability

### Root Cause

TiDB's `SELECT ... FOR UPDATE` can only acquire locks on rows that currently exist in the table. When a transaction reads a non-existent row (returning nil/null), no lock is acquired. This follows MySQL behavior but creates a gap:

- Transaction A reads non-existent key K1 with `SELECT ... FOR UPDATE` → no lock acquired
- Transaction B later writes to key K1
- Transaction C reads non-existent key K2 → no lock
- Transaction A writes to key K2
- Result: Cycles become possible when the first write to a key is involved

This is fundamentally different from snapshot isolation (SI), where such cycles are expected. With pessimistic locking and `SELECT ... FOR UPDATE`, cycles should not occur.

## Solution Approach

Implement gap locking semantics for `SELECT ... FOR UPDATE` on non-existent rows to ensure serializable isolation:

### 1) Gap Lock Acquisition

When `SELECT ... FOR UPDATE` encounters a range scan that returns no rows or returns nil for a specific key lookup:
- Acquire implicit gap lock on the searched range
- Gap lock prevents other transactions from inserting/writing in that gap
- Blocks phantom reads and write skew cycles

### 2) Lock Compatibility

Gap locks must be compatible with existing row locks:
- S (Shared) + gap lock: compatible
- X (Exclusive) + gap lock: compatible during same transaction
- X (Exclusive) from other transaction: blocked
- Insert into locked gap: blocked

### 3) Pessimistic Locking Enhancements

Modify pessimistic lock manager (`pkg/kv/pessimistic.go`, `executor/seqscan.go`):
- Track gap locks separately from row locks
- Extend lock table to include (table_id, index_id, key_range) tuples
- On transaction commit: release all gap locks
- On conflict detection: return lock conflict error

### 4) Integration Points

**Executor Layer (`pkg/executor/`):**
- `SeqScanExec`: Acquire gap locks for non-existent rows in `SELECT ... FOR UPDATE`
- `IndexLookupExec`: Acquire gap locks when index range is exhausted
- `TableReaderExec`: Track accessed ranges for gap lock placement

**KV Layer (`pkg/kv/`):**
- `LockResolver`: Handle gap lock resolution
- `Lock`: Extend to support lock type `Gap`
- `TxnState`: Track gap locks in transaction context

**Storage Layer (`pkg/store/`):**
- Update lock info encoding to include gap lock types
- Ensure gap locks are stored with minimal overhead

## Files Changed

### Core Implementation
- `pkg/kv/pessimistic.go` - Gap lock acquisition and management
- `pkg/executor/seqscan.go` - Gap lock for sequential scans
- `pkg/executor/index_lookup.go` - Gap lock for index lookups
- `executor/chunk_reader.go` - Track ranges in chunk processing

### Lock Management
- `pkg/lock/lock_waiter.go` - Wait graph for gap locks
- `pkg/lock/lock_resolver.go` - Resolve gap lock conflicts
- `session/session.go` - Expose gap lock configuration

### Configuration
- `config/config.go` - Add `pessimistic-gap-lock` flag (default: true)
- `pkg/session/vars.go` - Transaction variable for gap lock behavior

### Tests
- `tests/realtikvtest/pessimistic/gap_lock_test.go` - Real TiKV gap lock tests
- `tests/integrationtest/isolation/write_skew_test.go` - Write skew prevention tests
- `pkg/executor/seqscan_test.go` - SeqScan gap lock unit tests
- `pkg/kv/pessimistic_test.go` - Pessimistic lock manager tests

### Documentation
- `docs/design/pessimistic-locking-gap-locks.md` - Design document
- `README.md` - Updated isolation level documentation

## Type of Change

- [x] Bug fix (prevents serializability violation)
- [ ] New feature
- [ ] Breaking change

## Isolation Level Impact

**Snapshot Isolation (SI):** No change—write skew cycles remain allowed.

**Repeatable Read (RR):** 
- Without `SELECT ... FOR UPDATE`: Behavior unchanged
- With `SELECT ... FOR UPDATE`: Now provides serializability guarantee

**Serializable:** Strengthened to actually prevent write skew cycles.

## How Has This Been Tested?

### Unit Tests
- `pkg/kv/pessimistic_test.go`: Gap lock acquisition and release
- `pkg/executor/seqscan_test.go`: Gap lock integration with executor
- `pkg/lock/lock_resolver_test.go`: Lock conflict resolution

### Integration Tests
- `tests/integrationtest/isolation/write_skew_test.go`: 
  - Reproduces issue #10444 examples
  - Verifies write skew is prevented after fix
  - Tests multiple transaction orderings

- `tests/integrationtest/isolation/anti_dependency_test.go`:
  - Reproduces anti-dependency cycle examples
  - Confirms cycles are prevented

### Real TiKV Tests
```bash
# Start TiKV playground
tidb-server &
tikv-server &

# Run gap lock tests
make test-gap-lock-realtikvtest

# Cleanup
pkill tidb-server
pkill tikv-server
```

### Jepsen Testing
Coordinate with Jepsen team to re-run historical test suite:
- `tidb-2.1.7` equivalent code
- `tidb-3.0.0-beta.1` equivalent code
- Verify no anti-dependency cycles detected

## Risk Assessment

### Low Risk
- Feature is opt-in via configuration flag (default: enabled)
- Backward compatible with existing code
- No changes to public API

### Moderate Risk
- **Performance:** Gap locks add overhead to lock metadata per transaction
  - Mitigation: Index gap locks by (table_id, index_id) for O(log n) lookup
  - Mitigation: Batch release gap locks on commit
- **Compatibility:** Tighter locking may cause deadlocks in poorly-written applications
  - Mitigation: Documented in release notes
  - Mitigation: Users can disable with config flag if needed

### Testing Requirements
- [ ] All unit tests pass (`make test`)
- [ ] All integration tests pass (`make test-integration`)
- [ ] Real TiKV tests pass (10+ minutes runtime)
- [ ] No regression in existing pessimistic lock tests
- [ ] Write skew prevention verified (new test suite)

## Rollback Plan

If regressions occur:

1. **Disable by default:**
   ```
   [txn]
   pessimistic-gap-lock = false
   ```

2. **Revert commits:**
   - Remove gap lock implementation from `pkg/kv/pessimistic.go`
   - Remove gap lock acquisition from `pkg/executor/seqscan.go`
   - Remove gap lock tests

3. **Revert config changes:**
   - Remove `pessimistic-gap-lock` flag from `config/config.go`

## Verification Checklist

- [ ] Gap lock mechanism implemented in pessimistic lock manager
- [ ] Gap locks acquired for non-existent row reads in `SELECT ... FOR UPDATE`
- [ ] Lock conflict detection prevents write skew cycles
- [ ] Configuration flag works (`pessimistic-gap-lock`)
- [ ] Unit tests added and passing
- [ ] Integration tests added and passing
- [ ] Real TiKV tests added and passing
- [ ] Write skew test cases from #10444 are now prevented
- [ ] Anti-dependency cycles from #10444 are now prevented
- [ ] No regressions in existing pessimistic locking tests
- [ ] Performance impact is acceptable (< 5% overhead)
- [ ] Documentation updated
- [ ] Code follows TiDB conventions and style
- [ ] PR ready for review

## References

- Issue #10444: Anti-dependency cycles & write skew with `SELECT ... FOR UPDATE`
- Jepsen analysis: http://jepsen.io/analyses/tidb
- MySQL `SELECT ... FOR UPDATE` docs
- PostgreSQL `FOR UPDATE` locking semantics
- SQL Isolation Levels and Anomalies: ANSI SQL standard

## Notes

This fix aligns TiDB's behavior with stated guarantees about `SELECT ... FOR UPDATE` providing serializability. It ensures that read-lock-based pessimistic locking actually prevents write skew and anti-dependency cycles, not just read skew.

The implementation preserves backward compatibility and allows users to opt out via configuration if needed, while making serializability the default behavior for applications using pessimistic locking.
