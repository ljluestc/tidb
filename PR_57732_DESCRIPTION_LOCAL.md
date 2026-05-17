### What problem does this PR solve?
Issue Number: close #57732

`ADMIN CHECKSUM TABLE` returns incorrect aggregate results in unistore for hash-partitioned tables.

For a table with one inserted row:
- expected `Total_kvs = 1`
- actual result reports `Total_kvs = 17` and an incorrect checksum value

This makes checksum output unreliable and can mislead consistency verification.

### What changed and how does it work?
This PR replaces unistore’s placeholder checksum response with real KV-based checksum aggregation.

Main changes:
1. In `pkg/store/mockstore/unistore/cophandler/cop_handler.go`, `handleCopChecksumRequest` now:
   - decodes `tipb.ChecksumRequest`
   - validates checksum algorithm (`ChecksumAlgorithm_Crc64_Xor`)
   - extracts effective ranges from coprocessor request key ranges
   - scans KV pairs in those ranges with `dbReader.Scan(...)`
2. Aggregate checksum statistics from scanned KV data:
   - `Checksum`: CRC64-ECMA over key/value pairs with XOR accumulation
   - `TotalKvs`: number of scanned KV pairs
   - `TotalBytes`: accumulated `len(key) + len(value)`
3. Add regression coverage in `pkg/executor/checksum_test.go`:
   - update `TestChecksum` to assert data-based `Total_kvs` for partitioned table + index
   - add `TestChecksumHashPartitionSingleRow` to cover issue repro and assert `Total_kvs = 1`

### Check List
Tests <!-- At least one of them must be included. -->

- [x] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
Automated validation commands:
  > - `./tools/check/failpoint-go-test.sh pkg/executor -run 'TestChecksum|TestChecksumHashPartitionSingleRow' -count=1`
  > - `./tools/check/failpoint-go-test.sh pkg/store/mockstore/unistore/cophandler -run TestNonExistent -count=1`

Manual test steps (optional):
  > - `create table tb(id int primary key) partition by hash (id) partitions 16;`
  > - `insert into tb values(1);`
  > - `admin checksum table tb;`
  > - Verify `Total_kvs` is `1`.
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.
Side effects

- [ ] Performance regression: Consumes more CPU
- [ ] Performance regression: Consumes more Memory
- [ ] Breaking backward compatibility

Documentation
- [x] Affects user behaviors
- [ ] Contains syntax changes
- [ ] Contains variable changes
- [ ] Contains experimental features
- [ ] Changes MySQL compatibility

### Release note
```release-note
Fix incorrect `ADMIN CHECKSUM TABLE` result in unistore for hash-partitioned tables by using KV-based checksum aggregation instead of placeholder response values.
```
