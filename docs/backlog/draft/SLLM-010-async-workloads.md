# Async workloads

**Code:** SLLM-010
**Status:** Draft
**Persona:** Application developer running long model tasks

## Story

As an application developer, I want optional async workloads, so that long-running document, batch, or pipeline tasks do not need to hold an HTTP request open.

## Scope Ideas

- Job submission endpoint.
- Job status endpoint.
- Cancellation.
- Queue backend.
- Retry policy.
- Result retention policy.

## Acceptance Ideas

- Synchronous chat remains the simple default.
- Async jobs can survive client disconnects.
- Job metadata does not expose sensitive prompt content.
