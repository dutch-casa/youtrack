# 0030 Raw Response Files

## Status

Accepted.

## Context

`yt raw` is the completeness boundary for YouTrack REST operations that are not yet first-class commands. Some REST responses are not JSON text, especially attachment downloads or other long-tail endpoints that return exact file bytes.

Before this decision, the client named raw responses as `json.RawMessage`, and the CLI always printed them with `Fprintln`. That was acceptable for JSON APIs, but it made the raw boundary dishonest for byte responses because stdout printing adds a newline and implies text.

## Decision

Make `youtrack.Client.Raw` return `[]byte`.

Write raw response bytes exactly to stdout when no output file is set. Add `yt raw --output-file/-o PATH` to write raw response bytes exactly with mode `0600` when callers want file-safe downloads or do not want binary bytes in the terminal.

## Consequences

Agents can use the same raw escape hatch for JSON endpoints and byte-producing endpoints.

The YouTrack client surface now exposes the response fact it actually owns: raw bytes from HTTP, not JSON validity.

The CLI no longer appends a newline to `yt raw` stdout. Text and JSON callers that want a trailing newline can add it explicitly in their shell or consuming program.

## Rejected Alternatives

- Keep returning `json.RawMessage`: compatible by accident, but misleading and too narrow for a raw HTTP boundary.
- Keep stdout as `Fprintln(string(data))`: convenient for casual text responses, but it mutates the response and makes the raw boundary dishonest.
