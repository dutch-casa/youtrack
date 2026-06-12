# 0028 TUI Prompt Lifecycle

## Status

Accepted.

## Context

The TUI has several focused prompts: YouTrack command application, quick comments, quick work items, and query refinement. Each prompt needs the same lifecycle operations: construct with stable terminal defaults, open with focus, close by clearing transient text, and clear action-level errors before a new prompt starts.

Keeping that knowledge inline in `model.Update` made every new prompt widen the top-level state machine with repeated reset, focus, blur, and error-clearing code.

## Decision

Move prompt construction and lifecycle helpers into `internal/tui/prompts.go`.

The model still owns terminal state, selected issue state, loading state, and command dispatch. The prompt helper owns only prompt defaults and state transitions that are true for every focused text prompt.

Focused prompts keep keyboard ownership. Global navigation keys such as `r` are handled only after the prompt closes, so typing those characters inside a prompt remains literal input.

## Consequences

Adding a future prompt requires choosing the new mode and submit behavior, not copying reset/focus/error-clearing mechanics through the navigation handler.

The boundary stays intentionally local to `internal/tui`: it is not a generic text-input framework and does not hide Bubble Tea from the model.

## Rejected Alternatives

- Leave prompt lifecycle inline in `Update`: simple for one prompt, but the repetition now obscures which parts are mode-specific.
- Build a generic prompt registry: less duplicated code, but too much mechanism for four prompts and weaker type locality for submit behavior.
