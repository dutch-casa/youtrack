# 0009 TUI Evidence Browsing

## Status

Accepted

## Context

The CLI is agent-first, but the interactive mode should still be useful for humans browsing YouTrack. Comments and attachments are the main evidence attached to an issue. Before this decision, the TUI could inspect comments but not attachments, so a human had to leave the browser to understand issue evidence.

Interactive browsing must stay additive. It must not become the only way to perform core workflows, and it must reuse the typed client surface instead of learning YouTrack REST paths directly.

## Decision

Add an attachments pane to the TUI and cycle panes with `tab`:

- Details shows the selected issue summary, description, and custom fields.
- Comments lazy-loads issue comments.
- Attachments lazy-loads issue attachments through `internal/youtrack.Attachments`.

The TUI remains read-only for evidence. File upload stays in the non-interactive `yt attachments add` command where agents can call it predictably.

## Consequences

The human browser now covers the main evidence surfaces without widening the TUI into a second command system. Network loading stays explicit in Bubble Tea commands and cached per selected issue.
