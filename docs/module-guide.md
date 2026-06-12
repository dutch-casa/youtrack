# Module Guide

## `internal/auth`

Secret: This module hides how YouTrack credentials are discovered, prompted for, and stored locally.

Role: Callers ask for usable credentials. The module decides whether to use environment variables, an existing config file, or a terminal prompt, and it owns file permissions for persisted tokens.

## `internal/youtrack`

Secret: This module hides the YouTrack REST API wire shape.

Role: Callers use typed operations for common issue, comment, and user actions. The module owns endpoint paths, query fields, bearer-token headers, and API error decoding.

## `internal/ytcli`

Secret: This module hides the command-line contract users and agents call.

Role: It wires flags, subcommands, output defaults, and authentication into a stable CLI surface. JSON is the default because the primary caller is an agent.

## `internal/tui`

Secret: This module hides the terminal interaction state for browsing issues.

Role: It owns keyboard bindings, layout, selection state, and rendering for the lazygit-style interactive mode.
