# planner: support specially encoded unsigned handle for record key ordering (closes #10472)

## What problem does this PR solve?
For tables that use integer primary key as row handle, TiDB currently encodes both signed and unsigned handles with the same signed-int path (`codec.EncodeInt` over `int64`).

This preserves key order for signed handles, but does not preserve natural numeric order for unsigned handles across the full range. As a result, for unsigned values `a < b`, the encoded key order can be incorrect in boundary cases, which makes range semantics and planner assumptions harder to reason about.

## What is changed and how it works
This PR introduces a dedicated encoding path for unsigned row handles so that record key byte ordering is consistent with unsigned numeric ordering.

Main design points:
- Keep signed-handle behavior unchanged.
- For unsigned handles, encode using an unsigned-order-preserving representation.
- Ensure record key comparison in KV remains monotonic with handle comparison for both signed and unsigned handle types.

Likely touch points:
- Tablecodec record-key encode/decode helpers for row handles.
- Handle-aware planner/executor call sites that rely on key ordering behavior.
- Related tests for boundary values (`MaxInt64`, `MaxUint64`, cross-boundary comparisons).

## Why this approach
- Minimizes behavior change surface by only specializing unsigned-handle encoding.
- Makes key-order behavior explicit and type-correct.
- Reduces subtle bugs caused by relying on signed encoding for unsigned semantics.

## Risk assessment
- **Compatibility risk**: existing data encoding compatibility must be handled carefully for upgraded clusters.
- **Planner/executor risk**: range construction and handle comparisons must stay consistent.
- **Storage risk**: encode/decode symmetry and ordering guarantees must be validated with boundary tests.

## Test plan
- Unit tests for record key encoding/decoding:
  - signed monotonic ordering (regression)
  - unsigned monotonic ordering (new)
  - signed/unsigned boundary values
- Planner/executor tests for unsigned handle range behaviors.
- Backward compatibility checks for existing encoded rows where applicable.

## Checklist
- [ ] Code follows TiDB style and existing conventions.
- [ ] Tests added/updated for unsigned handle ordering and boundary values.
- [ ] Compatibility impact evaluated and documented.

## Issue reference
Closes #10472
