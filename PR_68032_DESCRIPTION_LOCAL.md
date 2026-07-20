Title: planner, expression: reuse one hidden generated column for identical expression indexes

What problem does this PR solve?

Issue Number: close #68032

Problem Summary:
Issue #67552 exposed a failure mode where one table can hold multiple expression indexes backed by different hidden generated columns, even when those indexes use the same expression. PR #67692 fixed the immediate planner resolution bug, but metadata duplication still exists at DDL time.

Today, creating expression indexes such as lower(name) repeatedly on the same table can create multiple hidden generated columns that are logically identical. This increases metadata redundancy, adds maintenance overhead, and keeps planner ambiguity surfaces wider than necessary.

Proposed Behavior:
When adding an expression index, if the table already has a hidden generated column for the same normalized expression and compatible derived metadata, reuse that existing hidden generated column instead of creating a new one.

Compatibility checks for reuse include normalized expression equivalence, compatible type-related properties, compatible collation or charset-related properties, and other index-required metadata compatibility checks used by DDL validation.

Safety and Lifecycle:
Reuse is scoped to new DDL operations in this stage.
Dropping one index must not drop a shared hidden generated column while other indexes still reference it.
Existing table metadata is not automatically rewritten in this stage.

What changed and how it works?
1) DDL expression-index creation path is updated to search for reusable hidden generated columns before allocating a new one.
2) Reuse decision is gated by normalized expression plus metadata compatibility checks.
3) Reference and lifecycle handling is adjusted so shared hidden generated columns are retained until no expression index references remain.
4) Planner and expression behavior continues to use exact generated-column identity, now benefiting from reduced duplicate metadata generation at source.

Why this approach?
It removes duplicated hidden generated columns for identical expression indexes, reduces planner ambiguity around duplicate virtual expressions, lowers metadata and maintenance overhead, and keeps rollout risk controlled by limiting scope to new DDL changes first.

Tests:
Added or updated DDL and expression-index tests to cover:
creating multiple identical expression indexes reuses one hidden generated column;
non-compatible metadata cases do not reuse;
dropping one of several indexes does not prematurely remove shared generated column;
planner and query behavior remains correct with reused generated columns.

Regression coverage ensures failure mode lineage from #67552 is protected together with DDL reuse behavior.

Backward compatibility:
No rewrite of historical metadata in this stage.
Existing tables continue to work as-is.
New DDLs progressively adopt deduplicated hidden generated column behavior.

Release note:
Reuse one hidden generated column for identical expression indexes on the same table when expression normalization and metadata compatibility checks pass.
