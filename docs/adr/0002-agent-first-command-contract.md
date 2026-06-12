# 0002 Agent-First Command Contract

## Status

Accepted

## Context

The default caller is an agent or script. That caller needs predictable output, no surprise prompts, and clear failures. Humans still need readable output and an interactive mode, but those are opt-in surfaces.

## Decision

The default CLI contract is:

- JSON output unless the user passes `--format table`.
- No interactive prompt unless stdin is a terminal and credentials are missing.
- Missing credentials in non-interactive mode return an actionable error naming both setup paths: `yt auth login --url ... --token ...` and `YOUTRACK_URL` / `YOUTRACK_TOKEN`.
- Mutating commands require explicit flags for required inputs.
- Long-tail API access goes through `yt raw METHOD PATH`, preserving full REST reach for agents while typed commands mature.

## Consequences

Every new command must be designed as a machine-stable interface first. Human convenience can be added through aliases, table output, and TUI actions, but it cannot change the default behavior.

The CLI can support any REST operation allowed by the token through `yt raw`, but typed first-class commands should be added for common workflows where the command name and validation reduce misuse.

## Rejected

- Interactive-first defaults: a prompt can deadlock an agent.
- Table output by default: humans like it, agents have to parse around it.
- A fully generic `request` command only: it provides reach but no pit of success for common workflows.
