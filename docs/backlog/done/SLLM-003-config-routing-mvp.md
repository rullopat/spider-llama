# Config-driven routing MVP

**Code:** SLLM-003
**Status:** Done
**Persona:** Application developer integrating model routing

## Story

As an application developer, I want to select models by alias, task, or requirements, so that client applications do not need to know the concrete backend model and node layout.

## Scope

- JSON config for nodes.
- JSON config for models.
- Model aliases.
- Capability, tag, node-tag, and context filters.
- Route preferences.
- Per-node in-memory concurrency limits.

## Acceptance

- Alias selection works with `alias:<name>`.
- `model: "auto"` can select by task route and requirements.
- Routing tests cover alias lookup, preferences, and concurrency limits.
