### What problem does this PR solve?
Issue Number: ref #2912

Problem Summary:
- When TiKV returns `StaleEpoch`, the target region may have just split and the request routing must be retried.
- Retrying through the generic backoff path can add unnecessary latency when the retried request still belongs to the same region.
- For rerouted requests that move to a different region after cache refresh, a short delay is still useful to reduce retry storms during split churn.

### What changed and how does it work?
- Adjust the TiKV request retry behavior for `StaleEpoch` in the region request path.
- If region lookup indicates the retried request still targets the same region, retry immediately without extra sleep.
- If the retried request is routed to a different region after region cache refresh, apply a short backoff (for example 50ms) before retrying.
- Keep existing retry guardrails intact (context cancellation, retry bounds, and retry error propagation).
- Add focused tests for both retry branches:
  - immediate retry for same-region stale-epoch handling
  - short-backoff retry for cross-region reroute handling

### Check List
Tests <!-- At least one of them must be included. -->

- [ ] Unit test
- [ ] Integration test
- [ ] Manual test (add detailed scripts or steps below)
- [ ] No need to test
  > - [ ] I checked and no code files have been changed.

Manual validation steps (to run when implementation is ready):
- Trigger region split and verify stale-epoch retry path does not sleep for same-region retries.
- Trigger region split with reroute and verify retry applies short backoff before resending.
- Confirm no retry regressions for non-`StaleEpoch` retryable errors.

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
Improve TiKV stale-epoch retry behavior by retrying immediately when the request remains in the same region, and using a short backoff when rerouting to a different region after split.
```
