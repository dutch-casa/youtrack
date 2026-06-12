# 0021 TUI Issue Pagination

## Status

Accepted

## Context

The TUI starts with a bounded issue list because YouTrack issue searches are paged by `$top` and `$skip`. Query refinement made the visible issue set dynamic, but the TUI still could not move beyond the first loaded page.

Leaving the TUI to relaunch with a different `--top` or dropping to non-interactive `issues list --skip` breaks the workspace feel.

## Decision

Use `n` and `p` for next and previous issue pages in interactive mode. Paging updates the existing issue-list options with `Skip += Top` or `Skip -= Top`, then reloads through the same typed `Issues` client operation used by launch, refresh, and query refinement.

Changing the query resets `Skip` to zero because a new query defines a new result set.

## Consequences

The TUI can browse large YouTrack result sets without adding a separate API surface.

Pagination remains page-oriented rather than infinite scrolling. This keeps the state explicit and matches the existing YouTrack REST pagination model.
