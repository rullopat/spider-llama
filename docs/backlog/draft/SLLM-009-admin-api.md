# Admin API for dynamic inventory

**Code:** SLLM-009
**Status:** Draft
**Persona:** Operator managing nodes without restarting the gateway

## Story

As an operator, I want to register and update nodes and models through an admin API, so that model inventory can change without editing config files and restarting the service.

## Scope Ideas

- Admin auth model.
- Create/update/disable nodes.
- Create/update/disable models.
- Persisted inventory.
- Config import/export.

## Acceptance Ideas

- Static config mode remains supported.
- Dynamic mode can add a node and route traffic to it.
- Disabled nodes and models are immediately excluded from routing.
