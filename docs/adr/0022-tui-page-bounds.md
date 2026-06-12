# 0022 TUI Page Bounds

## Status

Accepted

## Context

Interactive pagination uses YouTrack issue-list pages. The REST response gives the issues for the requested `$top` and `$skip`, but the TUI does not currently request a total count.

Advancing from a short final page would knowingly produce an empty or misleading next page.

## Decision

Enable next-page navigation only when the current page is full: `len(issues) == top`. A short page is treated as the end of the current result set. Previous-page navigation remains available whenever `skip > 0`.

Display the current page number in the issue list title instead of exposing raw skip arithmetic.

## Consequences

The TUI avoids walking past the known result set without adding an extra count request.

If YouTrack later exposes cheap total counts in this path, the TUI can replace this evidence rule with an explicit total.
