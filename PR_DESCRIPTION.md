### What is changed and how it works?

Fixes #28538

Case 1:
When the auto-generated ID exceeds the column type limit (e.g. `tinyint unsigned` max 255), the old behavior erroneously advanced the `LAST_INSERT_ID` to the overflowed value (e.g. 256) even though the insertion failed.

Case 2:
When `autoid` allocation happens for a batch of rows, failure in the middle would leave the session's temporary `LAST_INSERT_ID` in a dirty state.

This PR ensures that `LAST_INSERT_ID` is only updated when the allocated ID is valid within the column's range.

### Check List

Tests
- [x] Unit test
- [x] Integration test
- [x] Manual test (add detailed scripts or steps below)

### Manual Test Script

```sql
DROP TABLE IF EXISTS t1;
CREATE TABLE t1 (id tinyint unsigned not null auto_increment primary key);
INSERT INTO t1 VALUES (255);
-- Expect 255
SELECT last_insert_id();

-- This should fail with Duplicate entry '255', but NOT update last_insert_id to 256
INSERT INTO t1 VALUES (NULL);
-- Expect 255 (Previously returned 256)
SELECT last_insert_id();
```

### Side effects

- No backward compatibility issues.
- Improves MySQL compatibility (MySQL behaves as expected in this scenario).

### Related changes

- None
