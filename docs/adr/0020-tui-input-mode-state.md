# 0020 TUI Input Mode State

## Status

Accepted

## Context

The TUI has multiple prompt-like states:

- navigation, where normal keyboard bindings move through issues and panes
- command input, where `:` applies a YouTrack command to the selected issue
- query input, where `/` refines the visible issue list

Representing command and query prompts as separate booleans makes impossible UI states representable, such as command and query input both being active.

## Decision

Represent prompt focus with a single closed input-mode value: navigation, command, or query. Bubble Tea update handling dispatches to the active prompt before considering navigation keys.

## Consequences

Prompt focus is mutually exclusive by representation. Future prompt modes must extend the same state machine instead of adding another independent boolean.

Keys typed while a prompt is active belong to that prompt. For example, `/` inside command input is command text, not a query-mode transition.
