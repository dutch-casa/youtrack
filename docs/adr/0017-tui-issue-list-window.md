# 0017 TUI Issue List Window

## Status

Accepted

## Context

The interactive YouTrack workspace can load more issues than the left pane can display. Rendering the full issue list and relying on terminal clipping lets keyboard selection move outside the visible rows, which breaks the lazygit-style browsing contract.

## Decision

The TUI derives the visible issue-list window from the selected issue and available pane height. Selection remains the single navigation state for the left pane; moving with `j/k`, arrows, `g`, or `G` shifts the rendered window enough to keep the selected issue visible.

## Consequences

Agents and humans keep one issue-selection concept across list and detail panes. The list does not need a second viewport focus mode until it gains independent list-only scrolling or filtering behavior.
