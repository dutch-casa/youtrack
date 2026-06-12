# Module Guide

See also:

- [0001 Knowledge Boundaries](adr/0001-knowledge-boundaries.md)
- [0002 Agent-First Command Contract](adr/0002-agent-first-command-contract.md)
- [0003 Permanent-Token Authentication](adr/0003-permanent-token-auth.md)
- [0004 Charm TUI as Optional Human Mode](adr/0004-charm-tui-as-optional-human-mode.md)
- [0005 Command API as Workflow Bridge](adr/0005-command-api-as-workflow-bridge.md)
- [0006 Centralized Text Input](adr/0006-centralized-text-input.md)
- [0007 Typed Time Tracking](adr/0007-typed-time-tracking.md)
- [0008 Typed Attachments](adr/0008-typed-attachments.md)
- [0009 TUI Evidence Browsing](adr/0009-tui-evidence-browsing.md)
- [0010 Typed Activity History](adr/0010-typed-activity-history.md)
- [0011 Typed Link Browsing](adr/0011-typed-link-browsing.md)
- [0012 TUI Workspace Structure](adr/0012-tui-workspace-structure.md)
- [0013 Raw Body Sources](adr/0013-raw-body-sources.md)
- [0014 TUI Command Actions](adr/0014-tui-command-actions.md)
- [0015 Charm Text Input](adr/0015-charm-text-input.md)
- [0016 Scrollable TUI Panes](adr/0016-scrollable-tui-panes.md)
- [0017 TUI Issue List Window](adr/0017-tui-issue-list-window.md)
- [0018 Raw Request Content Type](adr/0018-raw-request-content-type.md)
- [0019 TUI Query Refinement](adr/0019-tui-query-refinement.md)
- [0020 TUI Input Mode State](adr/0020-tui-input-mode-state.md)
- [0021 TUI Issue Pagination](adr/0021-tui-issue-pagination.md)
- [0022 TUI Page Bounds](adr/0022-tui-page-bounds.md)
- [0023 Raw Request Headers](adr/0023-raw-request-headers.md)
- [0024 TUI Field Projection](adr/0024-tui-field-projection.md)
- [0025 TUI Quick Comments](adr/0025-tui-quick-comments.md)
- [0026 TUI Work Item Browsing](adr/0026-tui-work-item-browsing.md)
- [0027 TUI Quick Work Items](adr/0027-tui-quick-work-items.md)
- [0028 TUI Prompt Lifecycle](adr/0028-tui-prompt-lifecycle.md)

## `internal/auth`

Secret: This module hides how YouTrack credentials are discovered, prompted for, and stored locally.

Role: Callers ask for usable credentials. The module decides whether to use environment variables, an existing config file, or a terminal prompt, and it owns file permissions for persisted tokens.

## `internal/youtrack`

Secret: This module hides the YouTrack REST API wire shape.

Role: Callers use typed operations for common issue, comment, attachment, link, activity-history, and user actions. The module owns endpoint paths, query fields, multipart upload shape, activity category defaults, bearer-token headers, and API error decoding.

## `internal/ytcli`

Secret: This module hides the command-line contract users and agents call.

Role: It wires flags, subcommands, output defaults, and authentication into a stable CLI surface. JSON is the default because the primary caller is an agent.

## `internal/textinput`

Secret: This module hides how command text is resolved from literal flags, files, and stdin.

Role: Commands ask it to resolve one named text value. It enforces mutual exclusion across text sources and preserves the distinction between omitted text and an explicitly empty literal.

## `internal/tui`

Secret: This module hides the terminal interaction state for browsing issues.

Role: It owns keyboard bindings, layout, selection state, lazy-loaded issue panes, prompt modes for human issue actions, and rendering for the lazygit-style interactive mode. Model/update code owns terminal state transitions, prompt lifecycle code owns text prompt defaults and focus/clear behavior, view files own pane rendering, field projection owns which YouTrack custom fields become first-screen issue signals, formatting code owns terminal-safe text projection, and styles own the restrained YouTrack-like terminal skin.
