# 0035 Terminal Markdown and Help Rendering

## Status

Accepted

## Context

YouTrack long-form text appears in issue descriptions, knowledge-base articles, comments, project descriptions, and work item notes. Rendering those fields as raw Markdown makes the TUI feel unlike YouTrack and makes tables, lists, code blocks, and HTML entities harder to read.

The footer also had hand-written help strings. That duplicated keybinding knowledge already owned by the TUI update loop and made width-aware truncation a local formatting problem.

## Decision

Keep terminal text rendering inside `internal/tui`:

- Normalize terminal text before display, including HTML entities such as `&nbsp;`.
- Feed long-form Markdown through Charm Glamour before it enters the viewport.
- Treat knowledge-base article bodies and project descriptions as Markdown source, not pre-styled ANSI strings.
- Render the navigation footer through Bubbles `help` and `key` bindings instead of hand-written help prose.

Do not move Markdown rendering into `internal/youtrack`; the API client owns wire shape, not terminal presentation.

## Consequences

Markdown tables, lists, headings, and code blocks render consistently across the TUI. The resource substrate keeps one place for terminal-friendly text projection, while the client and non-interactive CLI outputs remain machine-oriented.

The footer becomes a generated view over keybinding metadata. Future keybinding changes should update the keymap rather than editing prose in multiple places.

## Rejected

- Use raw Markdown everywhere: simple, but it fails the terminal workspace goal.
- Render Markdown in `internal/youtrack`: leaks terminal presentation into the REST boundary.
- Replace the entire list/detail workspace with Bubbles `list` immediately: potentially useful later, but too much behavioral churn for this rendering change.
