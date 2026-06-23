# Security

**Status:** Draft v0.1
**Last updated:** 2026-06-23

## Threat Model

`spider-llama` is intended for trusted applications calling private or self-hosted model capacity. It may carry sensitive prompts and generated outputs. It should be treated as infrastructure, not as a public unauthenticated API.

## MVP Controls

- Bearer-token auth for non-health endpoints.
- Request body size limit.
- Per-node concurrency limits.
- Backend API keys can be loaded from environment variables.
- Gateway-only routing fields are stripped before forwarding chat requests.

## Deployment Guidance

- Bind backend inference servers to `127.0.0.1` when they are on the same machine.
- Prefer a private network or secure tunnel for multi-machine deployments.
- Do not expose raw inference backends directly to the public internet.
- Use firewall rules so only trusted clients can reach `spider-llama`.
- Rotate bearer tokens periodically and whenever a client host is compromised.
- Keep raw prompts and outputs out of logs by default.

## Sensitive Data Handling

The gateway should log operational metadata, not content:

- request ID
- selected route
- selected node/model
- latency
- status code
- error class

Avoid logging:

- raw prompts
- uploaded documents
- generated content
- authorization headers
- backend API keys

## Future Hardening

- Constant-time token comparison.
- HMAC request signing.
- mTLS between clients and the gateway.
- Per-client tokens and scopes.
- Rate limits per client.
- Audit log redaction tests.
- Config validation that rejects public backend URLs unless explicitly allowed.
