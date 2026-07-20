Title: feat: add Strimzi KafkaMirrorMaker CRD custom health checks

What problem does this PR solve?

Issue Number: close #5048

Problem Summary:
Argo CD currently lacks a custom health check for the Strimzi KafkaMirrorMaker custom resource, which can cause KafkaMirrorMaker objects to be reported as Healthy before their dependent underlying resources (such as deployments and related configurations) are actually ready.

In environments that rely on sync waves and application dependencies, this premature Healthy status can trigger downstream deployments too early and break intended deployment ordering.

What changed and how does it work?

1) Add a custom health check implementation for the KafkaMirrorMaker CRD, aligned with the existing pattern used for other Strimzi-managed resources.
2) Introduce Lua-based health check logic that evaluates KafkaMirrorMaker status conditions and readiness signals to distinguish Progressing, Healthy, and Degraded states more accurately.
3) Reuse the same design approach proven in the earlier Strimzi custom health check work (for example the pattern used in #3684), while adapting field checks to KafkaMirrorMaker-specific status details.
4) Add tests to validate expected health-state transitions and to prevent regressions in readiness evaluation.

Check List

Tests

- [ ] Unit test
- [x] Integration test
- [ ] Manual test (add detailed scripts or steps below)
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

Test details:
- Add test cases for KafkaMirrorMaker resources transitioning through initial creation, progressing state, healthy state, and failure/degraded conditions.
- Verify resources are not marked Healthy before required underlying components are actually ready.
- Verify dependent Argo CD applications can rely on corrected health state ordering in sync-wave based deployment flows.

Release note

Add custom health checks for Strimzi KafkaMirrorMaker resources so Argo CD reports accurate progressing/healthy/degraded states and improves sync-wave deployment ordering.
