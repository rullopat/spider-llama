# Go gateway foundation

**Code:** SLLM-001
**Status:** Done
**Persona:** Operator deploying a small private LLM gateway

## Story

As an operator, I want `spider-llama` to build as a single Go service, so that it is easy to run on a local or private machine without a large runtime stack.

## Scope

- Go module.
- `cmd/spider-llama` executable.
- Config loading.
- HTTP server startup and graceful shutdown.
- Unit-testable internal packages.

## Acceptance

- `go test ./...` passes.
- `go build -o bin/spider-llama ./cmd/spider-llama` creates a runnable binary.
- The service starts with the example config.
