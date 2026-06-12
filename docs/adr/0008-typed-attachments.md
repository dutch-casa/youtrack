# 0008 Typed Attachments

## Status

Accepted

## Context

Attaching files is a normal human YouTrack workflow and a common agent workflow for screenshots, logs, traces, and generated reports. YouTrack exposes issue attachment upload through `POST /api/issues/{issueID}/attachments` with multipart form data using repeated `upload` fields.

That multipart shape is a YouTrack wire detail. If callers have to build it themselves, every command and future integration learns the same fragile fact.

## Decision

Add a typed attachment surface:

- `internal/youtrack.Attachments` lists issue attachments with explicit fields.
- `internal/youtrack.UploadAttachments` accepts issue ID plus named readers and owns multipart encoding.
- `yt attachments list ISSUE` and `yt attachments add ISSUE --file PATH` expose the common workflow.
- `--file` can be repeated so one command can upload several attachments.

The CLI owns local filesystem selection. The YouTrack client owns endpoint paths, field selectors, and multipart form construction.

## Consequences

Agents get a stable non-interactive command for a core evidence-sharing workflow. The TUI can reuse the same client surface later without learning multipart details.

The current upload implementation buffers multipart bodies in memory. That keeps the first typed surface simple and testable; if large uploads become a real workload, the streaming implementation can change inside `internal/youtrack` without changing the command contract.
