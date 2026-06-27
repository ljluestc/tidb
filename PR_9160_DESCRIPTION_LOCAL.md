### What problem does this PR solve?
Issue Number: ref #9160

Problem Summary:
- MySQL 8.0 no longer allows `GRANT` to implicitly create users; behavior is equivalent to always enforcing `NO_AUTO_CREATE_USER`.
- TiDB historically allows `GRANT` to create missing users, which creates a compatibility gap for MySQL 8.0 strict behavior.
- We need a compatibility switch so deployments can opt into MySQL 8.0 semantics via `tidb_strict_compatibility_80`.

### What changed and how does it work?
- Add strict compatibility behavior in the privilege execution path for `GRANT`.
- When `tidb_strict_compatibility_80` is enabled and a `GRANT` target user does not exist, return MySQL-compatible error behavior instead of creating the user implicitly.
- Keep existing TiDB behavior unchanged when `tidb_strict_compatibility_80` is disabled.
- Add test coverage for:
  - `GRANT` to non-existent user with strict compatibility enabled (must fail)
  - `GRANT` to non-existent user with strict compatibility disabled (existing behavior preserved)
  - session/global scope interactions of `tidb_strict_compatibility_80`
- Update related compatibility documentation for the variable-controlled behavior.

### Check List
Tests <!-- At least one of them must be included. -->

- [ ] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.

Manual validation steps (to run when implementation is ready):
- `SET GLOBAL tidb_strict_compatibility_80 = ON;` then validate `GRANT` to non-existent user fails with expected MySQL-compatible error.
- `SET GLOBAL tidb_strict_compatibility_80 = OFF;` then validate existing TiDB compatibility behavior remains unchanged.
- Verify existing privilege workflows for already-created users are unaffected.

Side effects

- [ ] Performance regression: Consumes more CPU
- [ ] Performance regression: Consumes more Memory
- [ ] Breaking backward compatibility

Documentation

- [x] Affects user behaviors
- [ ] Contains syntax changes
- [x] Contains variable changes
- [ ] Contains experimental features
- [x] Changes MySQL compatibility

### Release note

```release-note
Add MySQL 8.0 compatible behavior for GRANT user creation under tidb_strict_compatibility_80: when enabled, GRANT no longer creates non-existent users implicitly.
```
