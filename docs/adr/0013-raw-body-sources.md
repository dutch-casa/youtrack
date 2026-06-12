# 0013 Raw Body Sources

## Status

Accepted

## Context

`yt raw` is the completeness escape hatch for agents. Typed commands should cover common workflows, but the raw command must still reach any REST operation allowed by the token.

Inline `--body` works for small JSON payloads. It is a poor interface for generated payloads, large updates, and shell-sensitive strings because callers have to escape JSON through the shell before the CLI can send it.

## Decision

Support three mutually exclusive raw body sources:

- `--body` for small inline JSON.
- `--body-file` for generated or checked-in JSON payloads.
- `--body-stdin` for pipeline-driven agents.

Use the existing `internal/textinput` resolver so raw body handling preserves the same one-source invariant as descriptions, comments, and command comments.

Validate body-source conflicts before resolving credentials, so local command-shape errors are reported before auth setup errors.

## Consequences

Agents can drive long-tail YouTrack REST operations without brittle shell escaping.

The raw surface remains intentionally low-level. It does not validate endpoint-specific payload schemas; typed commands should still be added when a common workflow has stable enough structure to make invalid calls harder to express.
