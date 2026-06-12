# 0018 Raw Request Content Type

## Status

Accepted

## Context

`yt raw` is the agent completeness escape hatch for YouTrack REST operations that are not yet worth a typed first-class command. It already supports inline, file, and stdin request bodies, but it assumed every body was JSON.

Most YouTrack REST write operations use JSON, so that default is correct. The raw boundary still needs to preserve lower-level HTTP facts when an endpoint expects a different media type.

## Decision

Represent raw REST calls inside `internal/youtrack` with a `RawRequest` struct containing method, path, content type, and body. Expose `yt raw --content-type`, defaulting to `application/json`.

Only requests with a body receive a `Content-Type` header.

## Consequences

Agents keep the simple JSON default while gaining enough control for non-JSON long-tail endpoints.

The raw surface remains intentionally low-level. Common workflows should still graduate into typed commands when the CLI can make their required fields and invariants harder to misuse.
