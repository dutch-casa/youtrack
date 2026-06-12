# 0015 Charm Text Input

## Status

Accepted

## Context

The TUI command prompt lets a human apply a YouTrack command to the selected issue. The first implementation kept the prompt state as a string and handled runes, backspace, enter, and cancel directly in the root Bubble Tea model.

That was enough for a narrow command path, but it made terminal text editing a local invention. Cursor movement, delete behavior, focus state, and rendering are terminal input concerns that Charm already models in Bubbles.

## Decision

Use `github.com/charmbracelet/bubbles/textinput` for the TUI command prompt.

Keep the root model responsible for when command mode opens, cancels, and submits. Let the Bubbles text input own prompt rendering and text editing while it is focused.

## Consequences

The command prompt now supports richer terminal editing behavior without expanding the TUI model's handwritten key handling.

The TUI depends on another Charm library, but this matches the product direction: interactive mode should be a Charm-based terminal workspace, while the default CLI remains non-interactive and dependency-light at the command surface.
