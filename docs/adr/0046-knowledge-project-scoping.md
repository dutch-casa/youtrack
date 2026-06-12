# 0046 Knowledge Project Scoping

## Status

Accepted

## Context

The non-interactive CLI supports `yt articles list --project PROJECT`, and the typed client already models project-scoped article listing. The TUI Knowledge Base section, however, listed only global articles, which makes large YouTrack instances hard to browse and leaves the TUI weaker than the CLI for a common article workflow.

Adding one Knowledge Base section per project would multiply top-level navigation and make project choice a structural concern. Treating project scope as a fuzzy local filter would be wrong because YouTrack exposes project-scoped article lists server-side.

## Decision

Reuse the existing project selector in the Knowledge Base section. Pressing `P` in Knowledge Base sets the resource project context and reloads articles through `ArticleListOptions.Project`.

The Knowledge Base list title includes the selected project. Resource pagination keeps using the same project context, and refresh reloads the first page of the scoped article list instead of clearing the scope.

Issue project filtering remains separate state. `projectFilter` continues to belong to issue search, while Knowledge Base project scope lives in resource context.

## Consequences

Users can browse project-specific Knowledge Base articles without leaving the TUI or falling back to raw REST calls.

The TUI keeps one Knowledge Base section and one resource substrate. Project scoping is a browsing context, not a new section or a duplicated article browser.

Future server-side article search can compose with this scope if the CLI adds a typed article search contract.

## Rejected

- Add project-specific Knowledge Base sections: creates unbounded navigation and repeats the same article UI.
- Use local fuzzy filtering for project scoping: filters only loaded pages and does not match the CLI/server-side article contract.
- Reuse issue `projectFilter`: couples issue search state to article browsing and makes switching sections surprising.
