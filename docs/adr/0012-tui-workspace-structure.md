# 0012 TUI Workspace Structure

## Status

Accepted

## Context

The interactive mode started as a compact issue browser. That was enough to validate Bubble Tea integration, but a terminal interface that feels like YouTrack needs more than an issue list and a detail pane. It needs issue details, comments, links, activity, attachments, and later issue actions without turning one file into the whole application.

The CLI remains agent-first and non-interactive by default. The TUI is an additive human workspace that must reuse the typed client surface instead of becoming a separate implementation.

## Decision

Structure `internal/tui` around the knowledge each file owns:

- `tui.go` owns Bubble Tea model state, update transitions, pane selection, and data-loading commands.
- `views.go` owns terminal layout and pane rendering.
- `format.go` owns terminal-safe projection of YouTrack values into readable text.
- `styles.go` owns the restrained Charm/Lip Gloss visual skin.

Add issue links and activity as lazy-loaded panes alongside details, comments, and attachments.

## Consequences

The interactive mode can grow toward a terminal YouTrack workspace without making every change edit one file.

The TUI still uses the same typed client operations as the non-interactive commands. This keeps human browsing and agent commands aligned, and preserves the raw/command escape hatches for operations that are not yet first-class panes or actions.
