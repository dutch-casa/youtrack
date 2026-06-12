# 0048 Help Desk Ticket Guidance

## Status

Accepted.

## Context

The Help Desk surface is issue-backed, but the TUI and CLI were not explicit about the split between help desk projects and ticket browsing. The TUI section reads as a project browser, while `yt helpdesk tickets` operates on a project key and uses the regular issue search contract.

Without clearer guidance, agents can waste time filtering the wrong surface or treat a regular project like a help desk project.

## Decision

Make the Help Desk contract explicit in both surfaces.

- The TUI Help Desk section labels itself as a project list, says `/` filters projects, and says `enter` opens tickets.
- Empty help desk ticket views explain that the loaded list is issue-backed and point to `yt helpdesk tickets PROJECT` or `yt issues list --query 'project: PROJECT ...'` as the next action.
- `yt helpdesk tickets PROJECT` validates that the project exists and is actually a help desk project before listing issues. If not, it returns a direct error that names the next command to use.

## Consequences

Agents get a clearer path from the Help Desk surface to the right follow-up command.

The CLI now fails fast on the wrong project type instead of quietly returning an empty ticket list, which is a better fit for command-driven workflows.

## Rejected

- Keep the old generic empty states: they leave the caller guessing which command to run next.
- Make the TUI track a separate help desk ticket model: too much structure for an issue-backed surface.
- Treat any project as a help desk project: that would blur the contract and hide a misuse.
