# 0026 TUI Work Item Browsing

## Status

Accepted.

## Context

Work items are YouTrack's time-tracking evidence for an issue. The CLI already models them as a typed non-interactive noun group with `yt work-items list` and `yt work-items add`.

The TUI aims to feel like YouTrack in the terminal. It can already browse comments, links, activity, and attachments, but it hides time-tracking evidence unless the user leaves the TUI and runs a separate command.

## Decision

Add a lazy-loaded `work` pane to the TUI tab cycle between activity and attachments.

The pane uses the existing typed `Client.WorkItems` operation, displays duration, type, author, date, and text, and keeps work-item REST path knowledge inside `internal/youtrack`.

## Consequences

Humans can inspect time spent while browsing an issue without changing modes. Agents keep the deterministic `yt work-items` command surface for creation and scripting.

The TUI gains another cached evidence pane, so query/page refreshes must clear work-item cache along with comments, links, activity, and attachments.

## Rejected Alternatives

- Add work-item creation to the TUI in the same change: useful, but it introduces duration/type/date input design and should be handled separately.
- Fold work items into the activity pane: convenient, but work items are their own YouTrack resource and deserve a stable pane like attachments.
