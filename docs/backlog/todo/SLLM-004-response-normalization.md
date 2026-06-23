# Response normalization

**Code:** SLLM-004
**Status:** Todo
**Persona:** Application developer consuming multiple model backends

## Story

As an application developer, I want consistent response shapes across providers, so that client code can handle text, JSON, and tool-style outputs predictably.

## Scope

- Define the gateway's normalized response contract.
- Preserve OpenAI-compatible responses where possible.
- Normalize known backend differences for tool calls and structured JSON.
- Add tests with representative provider responses.

## Acceptance

- Text responses preserve OpenAI-compatible fields.
- Structured JSON responses can be requested and validated.
- Tool-call-like responses have a consistent shape where supported.
- Backend-specific quirks are covered by tests.
