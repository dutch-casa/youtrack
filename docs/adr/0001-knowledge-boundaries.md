# 0001 Knowledge Boundaries

## Status

Accepted

## Context

The CLI needs to serve two callers:

- agents that need stable, non-interactive JSON commands
- humans that need a fast terminal browser for common YouTrack work

The tempting structure is a command pipeline split into handlers, services, repositories, and UI helpers. That would group code by when it runs, not by what it knows. YouTrack-specific change would then smear across every layer.

## Decision

Keep the repository split by knowledge:

- `internal/auth` owns credential discovery, prompting, validation, redaction, and secure local persistence.
- `internal/youtrack` owns YouTrack REST wire knowledge: paths, field selectors, bearer-token headers, request/response shapes, and API error decoding.
- `internal/ytcli` owns the public command contract: command names, flags, defaults, aliases, output mode, and dependency wiring.
- `internal/output` owns machine and human output representation.
- `internal/tui` owns terminal interaction state, keyboard bindings, layout, and rendering.

The command layer may orchestrate these modules, but it must not learn endpoint paths or token storage details. The TUI may consume typed client operations, but it must not become the only way to perform a workflow.

## Consequences

Adding first-class support for more YouTrack workflows should usually land in two places:

- `internal/youtrack` for the typed REST operation
- `internal/ytcli` for the command surface

TUI changes land in `internal/tui` unless they reveal a missing typed YouTrack operation. Raw endpoint access stays in `yt raw` for long-tail and admin operations so the first-class command surface can grow by actual usage rather than speculative completeness.

## Rejected

- Layer folders such as `handlers`, `services`, or `repositories`: these hide no domain decision and create lockstep edits.
- TUI-first architecture: this would make the human workflow the center and weaken the non-interactive agent contract.
- Code generation from the whole API as the primary interface: this would be wide and brittle before the useful command vocabulary is understood.
