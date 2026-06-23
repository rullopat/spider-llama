# Redis shared state

**Code:** SLLM-008
**Status:** Draft
**Persona:** Platform maintainer running more than one gateway instance

## Story

As a platform maintainer, I want optional Redis-backed state, so that multiple gateway instances can share health, concurrency, and circuit-breaker decisions.

## Scope Ideas

- Redis configuration.
- Shared node leases.
- Shared circuit breaker state.
- Health cache.
- Graceful fallback to stateless mode.

## Acceptance Ideas

- Stateless mode remains the default.
- Two gateway processes respect the same backend concurrency cap when Redis is enabled.
- Redis outages fail predictably and safely.
