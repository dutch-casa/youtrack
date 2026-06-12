# 0044 Agile Sprint Resource Drilldown

## Status

Accepted

## Context

The CLI exposes Agile boards and board sprints through typed commands. The TUI already lists Agile boards through the shared resource substrate, but it stopped at board metadata even though sprints are a normal part of YouTrack's Agile surface.

Adding a separate Sprint section would make users choose between boards and sprints before selecting a board. Keeping sprints only in board detail would hide a first-class list behind prose and would not support fuzzy filtering or mouse selection.

## Decision

Keep Agile as one top-level TUI section. The section initially lists boards; pressing `enter` on a board replaces the resource list with that board's sprints and changes the list title to `Sprints: BOARD`.

The sprint list reuses the existing resource substrate for list rendering, detail rendering, fuzzy filtering, browser opening, mouse selection, and scrolling. It does not advertise the board drill-down action once the list contains sprints.

Refreshing the Agile section reloads boards, which gives users a simple way back without adding a navigation stack.

## Consequences

Agile browsing gains board-to-sprint depth while preserving the small top-level section model.

The TUI client surface consumes the existing typed `Sprints` method instead of routing through raw REST calls.

Future sprint-specific actions can attach to sprint resources if there is a stable workflow to model, but the current change does not invent lifecycle operations beyond browsing.

## Rejected

- Add a separate top-level Sprints section: sprints require a board context, so the section would leak setup order into navigation.
- Embed all sprints in board detail only: loses list scanning, fuzzy search, and mouse selection.
- Treat a sprint like a board for repeated `enter`: misleading, because the selected resource no longer has an Agile board ID.
