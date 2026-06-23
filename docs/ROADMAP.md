# Roadmap

**Status:** Draft v0.1
**Last updated:** 2026-06-23

## 0.1.0 — llama.cpp-first Gateway MVP

Status: in progress.

Scope:

- Go single-binary service.
- Static JSON config.
- Local `llama.cpp` provider.
- Model aliases and capability/tag routing.
- Basic bearer auth.
- Health, nodes, models, route, and chat endpoints.
- Unit tests.

## 0.2.0 — Reliable Routing

Scope:

- Health-aware route-option ranking.
- Failure penalties and temporary node suppression.
- Better backend error mapping.
- Structured response normalization tests.
- Request ID propagation and structured logs.

## 0.3.0 — Provider Expansion

Scope:

- Additional OpenAI-compatible backend adapter.
- Ollama adapter if its behavior diverges enough to justify one.
- Provider capability declarations.
- Provider-specific config validation.

## 0.4.0 — Stateful Optional Mode

Scope:

- Redis-backed concurrency and health state.
- Distributed circuit breaker state.
- Multiple gateway instances sharing backend limits.
- Runtime stats endpoint.

## 0.5.0 — Long-Running Workloads

Scope:

- Async job API.
- Document/text pipeline primitives.
- OCR and extraction route shape.
- Job status and cancellation.

## 1.0.0 — Stable Public Gateway

Definition of done:

- Stable config schema.
- Stable HTTP API for chat, route inspection, model inventory, and health.
- At least two provider families supported.
- Release binaries for macOS, Linux, and Windows.
- Security documentation and hardening checklist.
- Versioned changelog and upgrade notes.
