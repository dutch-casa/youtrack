# 0011 Typed Link Browsing

## Status

Accepted

## Context

Issue links are core YouTrack navigation. Humans use them to understand dependencies, duplicates, subtasks, and related work before changing an issue. Agents need the same relationship graph without scraping text.

YouTrack exposes `/api/issues/{issueID}/links` as a collection of link buckets. Each bucket represents one issue link type and direction, then contains the linked issues for that bucket. Link writes require targeting the correct link bucket and issue relationship semantics.

## Decision

Add a typed client operation, `IssueLinks`, and a non-interactive command, `yt links list ISSUE`.

The command is read-only. Link mutation remains available through `yt commands apply` for YouTrack command-language workflows and `yt raw` for direct REST access until the CLI models link type discovery and write targeting explicitly.

## Consequences

Agents can inspect issue relationships as structured JSON or table output.

The typed surface avoids pretending that all link writes are interchangeable string operations. When typed link mutation is added, it should first expose the link type contract clearly enough that the obvious call is the correct call.
