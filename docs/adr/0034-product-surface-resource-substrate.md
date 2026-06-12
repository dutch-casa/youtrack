# 0034 Product Surface Resource Substrate

## Status

Accepted

## Context

The CLI must let agents do anything a human can do in YouTrack. The `raw` command already provides complete REST reach for long-tail endpoints, but first-class commands and the TUI were too issue-centric. YouTrack's product surface also includes knowledge base articles, help desk projects and tickets, agile boards, projects, and users.

Adding a separate hand-shaped TUI for every surface would duplicate list, selection, detail, filtering, mouse, and color behavior. Hiding all of those surfaces behind `raw` would technically preserve reach but would not make the CLI or terminal UI feel like YouTrack.

## Decision

Keep three boundaries:

- `internal/youtrack` owns typed REST knowledge for common YouTrack entities and preserves `raw` as the complete escape hatch.
- `internal/ytcli` exposes noun-based command groups for agent use.
- `internal/tui` owns a resource substrate that projects typed entities into searchable list/detail resources.

Issues remain the richest TUI section because they have first-class issue actions. Issue narrowing stays server-side through YouTrack search: `/` edits the issue query, `P` opens a project selector that applies a project filter, and `i` jumps to a specific issue ID. Other product areas use the shared resource substrate first: list, detail, section switching, mouse selection, browser opening, color treatment, and fzf-style live local search. Each section can later deepen with domain-specific actions without rebuilding navigation.

## Consequences

Common YouTrack surfaces are now easier to discover from the CLI and TUI, while uncommon or newly released YouTrack endpoints remain reachable through `yt raw`.

The TUI has one owner for terminal interaction state. REST endpoint paths and field selectors stay out of the TUI. CLI output remains JSON-first and non-interactive by default.

The resource substrate is intentionally small: it hides terminal browsing mechanics, not YouTrack REST semantics. If a product area needs richer behavior, that behavior should be added either as typed client methods plus CLI commands, or as a TUI section action that consumes those typed methods.
