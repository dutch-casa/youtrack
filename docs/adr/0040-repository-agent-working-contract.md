# 0040 Repository Agent Working Contract

## Status

Accepted

## Context

This project is explicitly built as an agent-first CLI. Future edits are likely to be made by coding agents as well as humans. The repository already records module boundaries and product decisions in ADRs, but the recurring working discipline also needs to be durable:

- use knowledge decomposition and interface design before widening surfaces
- preserve the non-interactive JSON command contract
- keep the Charm TUI additive
- commit atomically
- use commit messages as documentation
- avoid tool or assistant attribution in commits

Relying on chat history for this is too fragile. Future agents need a repo-local instruction file they can discover before editing.

## Decision

Add a root `AGENTS.md` working contract.

The file records the expected skills, module ownership boundaries, CLI/TUI product constraints, ADR discipline, atomic commit rules, and verification commands.

Keep the file concise and operational. It should tell an agent what must be preserved before it decides where to make a change.

## Consequences

The repository now carries the requested working memory in source control.

Future agent runs have a stable, visible contract for how to change the codebase and how to document those changes.

Changes to the working discipline should be reviewed like other project contract changes because they influence every future edit.

## Rejected

- Store the instruction only in chat memory: convenient for this session, but not durable in the repository.
- Put the instruction only in README: visible to users, but too noisy for product documentation and less targeted at coding agents.
- Encode the rule only in commit history: useful evidence, but it does not tell future agents what to do before their first commit.
