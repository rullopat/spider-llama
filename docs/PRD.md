# spider-llama Product Requirements

**Status:** Draft v0.1
**Last updated:** 2026-06-23

## Summary

`spider-llama` is an open-source LLM routing gateway for private and self-hosted model capacity. It gives applications a stable API for requesting models by alias, task, or capability while the gateway selects the concrete backend node and model.

The project starts with a local `llama.cpp` server backend and a stateless Go gateway. The architecture should remain reusable for other backends, multiple machines, and future stateful deployments.

## Goals

- Provide a small deployable gateway as a single Go executable.
- Route requests to models by alias, task, capability, tags, and context requirements.
- Prioritize private/self-hosted model backends while keeping the API easy for client applications.
- Start stateless, with clear extension points for distributed state, queues, and dynamic node management.
- Keep public documentation generic and reusable across projects.

## Non-Goals

- Training, fine-tuning, or model hosting inside the gateway process.
- Replacing backend inference servers such as `llama.cpp`, Ollama, vLLM, or OpenAI-compatible APIs.
- Owning client application workflows, billing, user accounts, or domain-specific business logic.
- Shipping a web admin UI in the MVP.

## Personas

- **Application developer:** wants a stable API and model aliases instead of hard-coding backend machines and concrete model names.
- **Home-lab/operator:** wants to expose local or private model capacity safely to trusted applications.
- **Platform maintainer:** wants a neutral routing layer that can grow from one local model to multiple nodes and providers.

## Core Use Cases

- Route an OpenAI-compatible chat request to a local `llama.cpp` instance.
- Let a client request `alias:light-text` instead of a backend model filename or internal deployment name.
- Select an available model matching requirements such as `text`, `json`, `tools`, `vision`, or `ocr`.
- Limit concurrent requests per node so one slow backend does not collapse the gateway.
- Expose health and model inventory endpoints for operators and client integration tests.

## MVP Requirements

- Single Go binary.
- JSON config file for nodes, models, aliases, capabilities, tags, and routes.
- Public `GET /health`.
- Authenticated `GET /v1/nodes`, `GET /v1/models`, `POST /v1/route`, and `POST /v1/chat/completions`.
- Bearer-token auth.
- Request body size limits.
- Per-node in-memory concurrency limits.
- `llama.cpp` provider using OpenAI-compatible `/v1/chat/completions`.
- Unit tests for config validation, routing, provider behavior, and HTTP handlers.

## Future Requirements

- Health-aware routing with latency and failure scoring.
- Multiple provider adapters.
- Response normalization for structured JSON and tool-call behavior across providers.
- Optional Redis-backed shared state for concurrency, health, and circuit breakers.
- Async job mode for long-running document and batch workloads.
- Admin API for dynamic node/model registration.
- Release builds for macOS, Linux, and Windows.
