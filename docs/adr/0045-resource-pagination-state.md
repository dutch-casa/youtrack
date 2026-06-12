# 0045 Resource Pagination State

## Status

Accepted

## Context

The TUI originally paged only issues. Other product surfaces used the shared resource substrate, but Knowledge Base, Help Desk, Agile boards, Agile sprints, Projects, and Users loaded only one page even though the typed YouTrack client and non-interactive commands support `top` and `skip`.

Reusing issue `opts.Skip` for resource pages would couple two different result streams. A user can start the TUI on an issue query page and then browse projects; those offsets should not affect each other.

## Decision

Keep issue pagination and resource pagination as separate TUI state:

- `opts.Skip` remains the issue-query offset.
- `resourceSkip` tracks the current page for the active resource list.
- `n/p` page the active list, whether that list is issues or resources.
- Resource section switches and refreshes reset `resourceSkip` to the first page.
- Agile sprint drill-down keeps the selected board ID as resource context so sprint pages request the correct board.

The resource substrate passes `resourceSkip` into typed YouTrack list calls rather than using raw REST paths.

## Consequences

Resource sections no longer silently stop at the first page.

Issue paging remains independent from product-surface browsing.

Local fuzzy filtering still filters the resources currently loaded in the active page. Server-side resource search remains a separate future decision because different YouTrack resources expose different query shapes.

## Rejected

- Reuse `opts.Skip` for every section: couples unrelated result streams and surprises users switching between Issues and other surfaces.
- Load all resource pages automatically: risks slow startup and unbounded requests on large self-hosted instances.
- Add section-specific pagination keys: makes the TUI harder to learn without adding power.
