# ADR-0001: Use Go for a single-binary gateway

**Status:** Accepted
**Date:** 2026-06-23

## Context

`spider-llama` needs to run as a small gateway on local and private machines. It should be easy to deploy without a large runtime stack, and it needs straightforward concurrency controls for backend model nodes.

## Decision

Build the gateway in Go.

## Consequences

- Users can build or download a single executable.
- The service can use the standard library for the MVP HTTP server, config parsing, and concurrency controls.
- Runtime deployment is simpler than a Python service with a virtual environment.
- Some ecosystem conveniences, such as rich YAML parsing or web frameworks, require explicit dependencies if needed later.
