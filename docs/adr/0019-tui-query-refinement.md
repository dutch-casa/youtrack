# 0019 TUI Query Refinement

## Status

Accepted

## Context

The interactive mode is intended to feel like a YouTrack workspace in the terminal. It could start with a `--query`, but changing that query required leaving the TUI and launching a new process.

That made browsing feel static and pushed a common human workflow outside the interactive boundary.

## Decision

Use `/` as an in-TUI YouTrack query prompt. The prompt starts with the current query, Enter reloads the issue list through the existing typed issue-list client operation, and Esc cancels without changing the active query.

Reloading a query resets issue selection to the first row and clears lazy-loaded pane caches because comments, links, activity, and attachments are keyed to the previous issue set.

## Consequences

The TUI can refine browsing scope without adding a second command-line entry point or a new YouTrack API surface.

The query prompt remains separate from `:` command mode. `/` changes the visible issue set; `:` applies a YouTrack command to the selected issue.
