Title: planner, expression: optimize range-scan plan building on unsigned integer columns

What problem does this PR solve?

Issue Number: close #11128

Problem Summary:
For predicates on unsigned integer columns, planner range derivation can be suboptimal when the constant is negative or non-integer negative (for example `a > -100` or `a > -100.1` with `a` as unsigned int).

Current behavior may keep a full table range scan `[-inf,+inf]` with an extra selection filter, instead of folding impossible negative bounds under unsigned semantics into tighter physical ranges.

Expected behavior is to derive the effective lower bound as `0` when the predicate semantics imply all valid unsigned values satisfy the comparison, so the scan range can be optimized to `[0,+inf]`.

What changed and how does it work?

1) Planner range-building logic for unsigned integer columns is adjusted to normalize negative constant comparisons into unsigned-valid bounds when safe.
2) For expressions such as `unsigned_col > negative_number`, the range builder now derives a direct table/index range starting from `0` instead of keeping a full-range scan plus filter.
3) The optimization path preserves correctness across integer and decimal negative constants by applying proper type-conversion and bound-check rules before simplification.
4) Cases that cannot be safely simplified continue using existing fallback behavior.

Check List

Tests

- [x] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.

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

Test details:
- Add planner/ranger unit coverage for unsigned integer range derivation with negative integer and negative decimal constants.
- Verify generated physical ranges are tightened (for representative examples, lower bound starts at 0).
- Verify unchanged behavior for non-simplifiable or type-incompatible cases.

Release note

Optimize planner range derivation for unsigned integer columns so predicates with negative constants can produce tighter scan ranges.
