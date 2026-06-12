# 0047 Top-Level Uninstall Command

## Status

Accepted.

## Context

The CLI already exposes `yt upgrade` for self-management, and the install script registers a paired `yt`/`youtrack` alias. Removing the tool by hand means remembering the installed binary name, the alias, and the saved auth config path.

That is a poor fit for an agent-first CLI. Self-management belongs in the command surface, and the uninstall path should reverse the same installation contract the tool already owns.

## Decision

Expose removal as `yt uninstall`.

The command resolves the installed binary the same way `yt upgrade` does, removes the installed binary, removes the paired alias when the name is `yt` or `youtrack`, and deletes the local auth config through `internal/auth`.

Keep the command under `internal/ytcli`. Uninstall is part of the CLI contract, not a separate auth or install subsystem.

## Consequences

Users can remove the CLI and its saved credentials without remembering install details or config paths.

The command stays machine-readable and additive. It does not change the default JSON contract or require interactive confirmation.

## Rejected

- Make users rerun the shell installer in reverse: works only if they remember the install details.
- Hide uninstall under a `self` namespace: too much structure for two maintenance commands.
- Leave uninstall to manual file deletion: discoverable only after the fact, and easy to do incompletely.
