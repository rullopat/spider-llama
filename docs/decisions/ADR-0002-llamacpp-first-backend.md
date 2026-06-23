# ADR-0002: Prioritize llama.cpp as the first backend

**Status:** Accepted
**Date:** 2026-06-23

## Context

The MVP should prove the gateway against a local inference server. `llama.cpp` is a good first target because `llama-server` can run locally and exposes an OpenAI-compatible chat endpoint.

## Decision

Implement `llamacpp` as the first provider adapter.

## Consequences

- The MVP can run entirely on one local machine.
- The initial API path can stay close to OpenAI-compatible `/v1/chat/completions`.
- Provider interfaces must remain generic so the project does not become a `llama.cpp`-only proxy.
