# 0030 Raw Response Files

## Status

Accepted.

## Context

`yt raw` is the completeness boundary for YouTrack REST operations that are not yet first-class commands. Some REST responses are not JSON text, especially attachment downloads or other long-tail endpoints that return exact file bytes.

Before this decision, the client named raw responses as `json.RawMessage`, and the CLI always printed them with `Fprintln`. That was acceptable for JSON APIs, but it made the raw boundary dishonest for byte responses because stdout printing adds a newline and implies text.

## Decision

Make `youtrack.Client.Raw` return `[]byte`.

Add `yt raw --output-file/-o PATH` to write raw response bytes exactly with mode `0600`. Keep existing stdout behavior unchanged when no output file is set so current JSON/text uses continue to work.

## Consequences

Agents can use the same raw escape hatch for JSON endpoints and byte-producing endpoints.

The YouTrack client surface now exposes the response fact it actually owns: raw bytes from HTTP, not JSON validity.

## Rejected Alternatives

- Keep returning `json.RawMessage`: compatible by accident, but misleading and too narrow for a raw HTTP boundary.
- Change stdout to exact byte streaming: more correct for binary output, but it would alter existing text behavior and can make terminal sessions unsafe for binary payloads.
