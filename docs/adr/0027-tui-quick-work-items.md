# 0027 TUI Quick Work Items

## Status

Accepted.

## Context

The TUI can now browse issue work items, and the non-interactive CLI already exposes precise work-item creation through `yt work-items add`.

Humans often need to log a simple amount of time while reviewing an issue. Leaving the TUI for the base CLI breaks the terminal workspace flow, but copying the full CLI flag surface into the TUI would add too much modal input design at once.

## Decision

Add a `w` keybinding that opens a focused quick-work prompt for the selected issue.

The prompt accepts a leading duration in a narrow grammar: `45m`, `1h`, `1h30m`, or separated parts such as `1h 30m`. Remaining text becomes the optional work-item text. Submit through the typed `Client.AddWorkItem` operation, invalidate work-item and activity caches for that issue, and reload the work pane only if the issue is still selected.

Keep advanced fields such as work item type, author, date, and notification muting in the non-interactive `yt work-items add` command until each has a deliberate terminal prompt design.

## Consequences

The TUI covers the common human time-entry path without weakening the agent-first CLI contract.

The duration parser becomes TUI-owned presentation/input knowledge. Its grammar is intentionally small so invalid or ambiguous duration text fails at the prompt instead of silently recording the wrong minutes.

## Rejected Alternatives

- Accept bare numbers as minutes: fast to type, but easy to confuse with issue text or other numeric notes.
- Add a full multi-field work-item form immediately: more complete, but it widens the TUI state machine before the simple path proves insufficient.
