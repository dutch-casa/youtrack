# 0010 Typed Activity History

## Status

Accepted

## Context

Agents need to understand how an issue reached its current state, not just read the current fields. In YouTrack, that evidence lives in the issue activity stream. The REST endpoint requires callers to provide activity category IDs and returns polymorphic activity items whose `added`, `removed`, and `target` values differ by activity type.

Without a typed command, agents must either memorize category IDs or parse raw activity payloads for common history inspection.

## Decision

Add a typed activity-history surface:

- `internal/youtrack.Activities` calls `/api/issues/{issueID}/activities`.
- The client requires at least one category because the YouTrack API requires it.
- `yt activities list ISSUE` uses a broad default category set for ordinary issue history.
- Repeated `--category` flags narrow the request to exactly those categories.
- `activity` and `history` aliases are stable entry points for the same command group.

`internal/youtrack.Activity.Summary` hides the polymorphic value extraction needed for human table output. JSON output still returns the full typed payload by default for agents.

## Consequences

Agents can inspect issue history non-interactively without dropping to `raw` for the common case. The raw escape hatch remains available for rare categories or fields not yet promoted into the typed surface.

The default category list is a product decision owned by the YouTrack client boundary. If YouTrack adds or renames categories, the change is local to `internal/youtrack` and the CLI command contract can stay stable.
