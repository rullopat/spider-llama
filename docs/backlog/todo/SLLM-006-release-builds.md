# Release builds

**Code:** SLLM-006
**Status:** Todo
**Persona:** User installing spider-llama without building from source

## Story

As a user, I want downloadable release binaries, so that I can install `spider-llama` without a local Go toolchain.

## Scope

- GitHub Actions release workflow.
- macOS, Linux, and Windows binaries.
- Checksums.
- Basic version command.
- Release notes template.

## Acceptance

- Tagging a release builds platform binaries.
- Checksums are attached to the release.
- Binary reports version, commit, and build date.
- README documents installation from release assets.
