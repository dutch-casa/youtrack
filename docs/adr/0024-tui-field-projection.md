# 0024 TUI Field Projection

## Status

Accepted.

## Context

YouTrack issue identity in the web UI is not just an ID and summary. Human triage depends on compact signals such as state, assignee, priority, type, project, and resolved status.

The REST client should continue to expose YouTrack issue data without deciding which custom fields deserve first-screen terminal treatment. That decision is presentation knowledge owned by the TUI.

## Decision

Move custom-field projection behind `internal/tui` helpers that parse YouTrack custom-field JSON into terminal-facing issue signals.

Render a compact metadata strip in the detail pane and show the selected issue state in the list row when available. Keep the full custom-field list in the details pane for evidence browsing.

## Consequences

The TUI feels more like YouTrack during browsing while the non-interactive client surface remains wire-shape focused.

Future changes to which fields are treated as high-signal terminal metadata should stay local to the TUI projection helper instead of spreading through view rendering or the REST client.

## Rejected Alternatives

- Add state, assignee, priority, and type as first-class fields on `youtrack.Issue`: convenient for this view, but it makes the REST client own a presentation ranking.
- Only keep the full custom-field list: complete, but it hides the issue signals humans scan for most often.
