# Health-aware routing

**Code:** SLLM-007
**Status:** Draft
**Persona:** Operator running multiple private model nodes

## Story

As an operator, I want routing to account for node health and recent failures, so that traffic moves away from unhealthy or overloaded backends automatically.

## Scope Ideas

- Background health probes.
- Failure counters.
- Temporary node suppression.
- Latency scoring.
- Route decision explanations.

## Acceptance Ideas

- Unhealthy nodes are excluded from automatic selection.
- Recently failing nodes are penalized.
- Manual alias selection returns a clear error when the target node is unavailable.
