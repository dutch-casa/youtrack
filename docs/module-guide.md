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

## `internal/auth`

Secret: This module hides how YouTrack credentials are discovered, prompted for, and stored locally.

Role: Callers ask for usable credentials. The module decides whether to use environment variables, an existing config file, or a terminal prompt, and it owns file permissions for persisted tokens.

## `internal/youtrack`

Secret: This module hides the YouTrack REST API wire shape.

Role: Callers use typed operations for common issue, comment, attachment, and user actions. The module owns endpoint paths, query fields, multipart upload shape, bearer-token headers, and API error decoding.

## `internal/ytcli`

Secret: This module hides the command-line contract users and agents call.

Role: It wires flags, subcommands, output defaults, and authentication into a stable CLI surface. JSON is the default because the primary caller is an agent.

## `internal/textinput`

Secret: This module hides how command text is resolved from literal flags, files, and stdin.

Role: Commands ask it to resolve one named text value. It enforces mutual exclusion across text sources and preserves the distinction between omitted text and an explicitly empty literal.

## `internal/tui`

Secret: This module hides the terminal interaction state for browsing issues.

Role: It owns keyboard bindings, layout, selection state, lazy-loaded evidence panes, and rendering for the lazygit-style interactive mode.
