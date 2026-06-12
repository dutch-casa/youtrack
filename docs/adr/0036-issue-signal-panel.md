# 0036 Issue Signal Panel

## Status

Accepted

## Context

Issue custom fields carry high-value browsing signals: project, state, assignee, priority, type, resolution, and project-specific fields. Showing them as plain lines appended after the Markdown description made those signals easy to miss and made the detail pane feel like raw output instead of a YouTrack workspace.

The TUI already owns terminal presentation, while `internal/youtrack` owns field decoding and custom-field value projection. Color meaning belongs with the TUI, not the REST client.

## Decision

Render issue fields in a dedicated `Signals` panel near the top of the issue detail pane:

- Project, state, assignee, priority, type, and resolved status are promoted first.
- Remaining custom fields still render in the panel after the promoted signals.
- Color is assigned by semantic kind: project, person, good, warning, bad, info, and default.
- The Markdown description remains below the signal panel and is no longer mixed with field rows.

## Consequences

Humans can scan the same high-value issue signals before reading the full description. The description viewport stays focused on Markdown content. Field styling changes stay local to `internal/tui`.

Agents still use the non-interactive JSON issue output for exact field values; the panel is a human browsing affordance only.

## Rejected

- Keep appending fields after the description: mechanically simple, but buries the most important browsing signals.
- Put field coloring in `internal/youtrack`: would leak terminal presentation into the REST boundary.
- Only show the promoted fields: cleaner, but hides project-specific fields that often carry workflow meaning.
