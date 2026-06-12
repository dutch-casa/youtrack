# 0004 Charm TUI as Optional Human Mode

## Status

Accepted

## Context

The CLI needs a lazygit-style browsing mode without making terminal UI state part of the core command contract. Charm Bubble Tea provides an idiomatic Go model/update/view loop for this kind of interface.

## Decision

Use Bubble Tea and Lip Gloss for an explicit `yt interactive` mode.

The TUI owns:

- loaded issue snapshots
- selected row
- terminal dimensions
- loading/error state
- keyboard bindings
- rendering

The TUI consumes typed YouTrack client operations. It does not own auth, output formats, or REST endpoint construction.

## Consequences

Human browsing can become richer without changing agent defaults. TUI actions should call the same typed operations exposed to non-interactive commands where possible.

Substantial layout changes need direct model tests for movement, refresh, bounds, and error rendering. Manual terminal checks are still needed for visual quality when layout changes.

## Rejected

- Building TUI state into the command layer: this would mix public command contracts with terminal interaction state.
- Making the TUI the default command: it would violate the agent-first contract.
- Hand-rolled terminal control: Bubble Tea already owns the event loop and terminal mechanics.
