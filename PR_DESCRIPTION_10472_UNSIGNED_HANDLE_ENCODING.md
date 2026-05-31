### What problem does this PR solve?
Issue Number: close #10472

Problem Summary:
TiDB currently encodes signed and unsigned integer row handles through the same signed path (`codec.EncodeInt` on `int64`) when building record keys. For unsigned handles, encoded key order may not match unsigned numeric order in boundary cases, which makes range semantics and planner assumptions harder to reason about.

### What changed and how does it work?
This PR description documents a feature direction to support dedicated unsigned-handle encoding for record keys so key order is monotonic with unsigned handle order.

Planned behavior:
1. Keep existing signed-handle encoding behavior unchanged.
2. Introduce an unsigned-order-preserving encoding path for unsigned handles.
3. Ensure encode/decode and comparison behavior are symmetric and monotonic for both signed and unsigned handles.
4. Cover signed/unsigned boundary cases with focused regression tests when code implementation is included.

### Check List
Tests <!-- At least one of them must be included. -->

- [ ] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [x] No need to test
  > - [x] I checked and no code files have been changed.
  > - This branch currently adds only a PR description markdown file.

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
