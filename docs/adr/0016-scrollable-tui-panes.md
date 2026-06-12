# 0016 Scrollable TUI Panes

## Status

Accepted

## Context

YouTrack issue descriptions, comments, activity, links, and attachment lists can exceed the visible terminal area. The TUI previously truncated pane content to the available height. That made the interface fast to render but unsuitable as a terminal YouTrack workspace because important issue context could be hidden with no way to reach it.

## Decision

Use `github.com/charmbracelet/bubbles/viewport` for the right-hand issue pane.

Keep pane functions responsible for producing full textual content. Keep the model responsible for viewport dimensions, scroll offset, and scroll key handling.

Reset scroll position when the selected issue or active pane changes. Preserve scroll position while the user pages through the current pane.

## Consequences

The TUI can browse long issue descriptions and evidence panes without losing content.

Pane rendering remains a content projection concern, while terminal scrolling stays in the Bubble Tea model where interaction state belongs.
