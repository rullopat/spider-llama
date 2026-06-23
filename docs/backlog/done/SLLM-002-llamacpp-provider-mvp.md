# llama.cpp provider MVP

**Code:** SLLM-002
**Status:** Done
**Persona:** Developer running a local `llama.cpp` model server

## Story

As a developer, I want `spider-llama` to proxy chat requests to `llama.cpp`, so that a local `llama-server` can be used behind the gateway.

## Scope

- Provider adapter for `llamacpp`.
- `GET /health` probe to the backend.
- `POST /v1/chat/completions` forwarding.
- Optional backend bearer token support.

## Acceptance

- Provider tests verify health and chat forwarding.
- Chat requests are sent to `/v1/chat/completions`.
- Backend authorization header is forwarded when configured.
