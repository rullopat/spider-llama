# Observability baseline

**Code:** SLLM-005
**Status:** Todo
**Persona:** Operator debugging private model routing

## Story

As an operator, I want structured operational logs and request IDs, so that routing failures and backend latency can be diagnosed without logging sensitive prompt content.

## Scope

- Request ID generation or propagation.
- Structured logs.
- Selected node/model logging.
- Latency and status logging.
- Explicit redaction policy.

## Acceptance

- Each request has a request ID.
- Logs include route decision metadata.
- Logs do not include raw prompts, generated content, auth headers, or backend secrets.
- Tests cover redaction helpers where applicable.
