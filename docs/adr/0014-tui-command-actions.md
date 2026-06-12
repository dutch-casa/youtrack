# 0014 TUI Command Actions

## Status

Accepted

## Context

The TUI can browse issue details, comments, links, activity, and attachments. That makes it useful for inspection, but YouTrack is also an action workspace: humans commonly change state, assign work, tag issues, and update links through YouTrack commands.

The CLI already exposes `yt commands apply` as the non-interactive bridge to YouTrack's command language. Duplicating every workflow as a TUI-specific form would widen the interface and create a second command system.

## Decision

Add a TUI command prompt opened with `:`.

The prompt applies the entered YouTrack command to the selected issue through the same typed client operation used by `yt commands apply`. It does not invent TUI-only workflow semantics.

After a successful command, refresh the issue list and clear cached relationship/history panes for the selected issue so follow-up browsing does not show stale state.

## Consequences

The interactive mode becomes a lightweight terminal YouTrack workspace instead of a read-only browser.

Humans can perform broad issue workflows from the TUI while agents keep using the non-interactive command surface. Typed TUI forms can still be added later for workflows where a dedicated interface makes misuse meaningfully harder.
