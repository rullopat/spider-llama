# Architecture

**Status:** Draft v0.1
**Last updated:** 2026-06-23

## System Shape

```text
client application
  -> spider-llama HTTP API
  -> router
  -> provider adapter
  -> model node
       - llama.cpp server
       - future OpenAI-compatible backends
       - future Ollama / vLLM / remote nodes
```

`spider-llama` is the routing and policy layer. It does not run model inference itself.

## Core Concepts

### Node

A node is a machine or endpoint that can serve one or more models.

Configured fields include:

- `id`
- `provider`
- `base_url`
- `tags`
- `max_concurrency`
- `timeout_seconds`

### Model

A model is a concrete model available through a node.

Configured fields include:

- `id`
- `backend_model`
- `node`
- `aliases`
- `capabilities`
- `tags`
- `context_tokens`
- `priority`

`id` is the public gateway-level model identity. `backend_model` is what the provider sends to the backend server.

### Alias

An alias is a stable public name for a model, such as:

- `light-text`
- `fast-json`
- `reasoning`
- `local-default`

Clients can request aliases with `model: "alias:light-text"`.

### Capability

Capabilities describe what a model can do:

- `text`
- `json`
- `tools`
- `vision`
- `ocr`
- `embeddings`
- `code`

Capabilities are declarative in the MVP. Provider-specific enforcement and probing can be added later.

### Route

A route maps a task name to requirements and preferences.

Example:

```json
{
  "analysis": {
    "require": {
      "capabilities": ["text", "json"],
      "tags": ["light"]
    },
    "prefer": {
      "node_tags": ["local"]
    }
  }
}
```

Task names are generic and application-defined.

## Routing Flow

For a chat request:

1. Parse `model`, `task`, and `requirements`.
2. If `model` is a concrete model or alias, select that model directly.
3. If `model` is empty or `auto`, load the task route.
4. Filter models by required capabilities, tags, node tags, context window, and enabled state.
5. Sort route options by priority and preferences.
6. Acquire the selected node's concurrency slot.
7. Rewrite gateway-only request fields out of the payload.
8. Send the request to the provider adapter.
9. Return the backend response with selected node/model headers.

## Provider Boundary

Provider adapters own backend-specific HTTP behavior:

- endpoint paths
- backend auth
- health checks
- timeout handling
- future response normalization

The current provider is `llamacpp`, which targets `llama-server`'s OpenAI-compatible API.

## State Model

The MVP is stateless apart from in-memory concurrency counters. This keeps deployment simple and makes the gateway useful as a local sidecar or private-network service.

Future state modules should be optional:

- Redis for shared health, concurrency, request leases, and circuit breaker state.
- SQLite/Postgres for audit logs, model inventory, or multi-tenant/project configs.
- A queue for async document and batch workloads.

## API Surface

Current:

- `GET /health`
- `GET /v1/nodes`
- `GET /v1/models`
- `POST /v1/route`
- `POST /v1/chat/completions`

The chat endpoint intentionally stays close to OpenAI-compatible request/response shapes.

## Deployment Model

MVP deployment:

```text
llama-server on 127.0.0.1:8080
spider-llama on 127.0.0.1:8088 or a private-network interface
trusted client app calls spider-llama with bearer auth
```

For remote private nodes, prefer a private network such as WireGuard/Tailscale or an equivalent tunnel. Avoid exposing raw inference backends directly to the public internet.
