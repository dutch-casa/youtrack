# 0005 Command API as Workflow Bridge

## Status

Accepted

## Context

YouTrack has many issue workflows that humans can perform from the UI: assigning issues, changing states, tagging, linking, starring/watching, and applying workflow-specific commands. Implementing each as a separate bespoke CLI command would make the surface wide before usage proves which workflows deserve first-class names.

YouTrack also exposes a command API that applies the same command-language style operations to issues.

## Decision

Expose YouTrack's command API as:

```sh
yt commands apply ISSUE --query 'State Fixed' --comment 'Fixed in main'
```

Keep `yt raw` as the escape hatch for any REST endpoint allowed by the token. Add first-class typed commands only when they make common work safer or more discoverable than either `commands apply` or `raw`.

## Consequences

Agents can perform broad issue workflow changes without waiting for a new binary release for every YouTrack operation. The CLI still has a path to polished commands for high-frequency work.

This is not the same as claiming every human UI affordance has a typed command. The capability ladder is:

1. `yt raw` can reach any REST endpoint permitted by the token.
2. `yt commands apply` can perform command-language issue workflows.
3. Typed commands provide safer, discoverable paths for common operations.
4. The TUI provides optional human browsing and should reuse typed operations as actions are added.

## Rejected

- Generating a huge command tree for the entire REST API up front: too wide and hard to stabilize.
- Only providing `raw`: complete reach, but poor ergonomics and weak validation for common workflows.
- Encoding every field update as a custom flag immediately: YouTrack custom fields are project-specific and are better handled first by command language and raw JSON until usage patterns are clear.
