# 0025 TUI Quick Comments

## Status

Accepted.

## Context

The CLI is agent-first and already exposes non-interactive comment creation through `yt comments add`. The TUI is the explicit human mode, and discussion is a common YouTrack issue workflow while browsing.

The TUI already has focused prompt modes for issue query refinement and YouTrack command application. Comment entry is the same kind of terminal interaction knowledge: it is human, selected-issue scoped, and should reuse the typed client instead of constructing raw REST requests.

## Decision

Add a `c` keybinding that opens a focused single-line quick-comment prompt for the selected issue.

Submit non-empty trimmed text through `Client.AddComment`, then switch to the comments pane, invalidate comments and activity caches for that issue, and reload comments. Keep long-form or file/stdin comment creation in the non-interactive `yt comments add` command.

## Consequences

Humans can participate in issue discussion without leaving the TUI, and agents keep the existing deterministic command surface.

The TUI client interface grows by one typed operation, but the REST path knowledge remains owned by `internal/youtrack`.

## Rejected Alternatives

- Use `yt raw` from the TUI: complete, but it leaks REST mechanics into terminal interaction state.
- Add a multi-line editor immediately: better for long comments, but it adds modal editing complexity before the quick discussion path is proven insufficient.
