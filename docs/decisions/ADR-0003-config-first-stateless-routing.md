# ADR-0003: Start with config-first stateless routing

**Status:** Accepted
**Date:** 2026-06-23

## Context

The gateway needs node/model routing, but it should not require a database, Redis, or an admin service for the first useful version.

## Decision

Use a static JSON config file for nodes, models, aliases, capabilities, tags, and routes. Keep runtime state in memory for the MVP.

## Consequences

- The service remains easy to understand and deploy.
- Config is reviewable in source control.
- Runtime inventory changes require a restart.
- Multi-instance deployments will need optional shared state later.
