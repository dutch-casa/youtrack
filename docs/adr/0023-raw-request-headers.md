# 0023 Raw Request Headers

## Status

Accepted.

## Context

`yt raw` is the completeness boundary for YouTrack REST operations that are not yet modeled as typed commands. Some long-tail endpoints require request headers beyond method, path, body, and content type.

The raw command still must not let callers bypass the CLI's credential model. Agents should be able to set endpoint-specific headers, but bearer authorization should remain owned by `internal/auth` and `internal/youtrack`.

## Decision

Add repeated `yt raw --header/-H 'Name: value'` flags and carry them through `youtrack.RawRequest.Headers`.

Reject `Authorization` headers at both the CLI parser and client boundary. Reject `Content-Type` headers and keep media type selection in the existing `--content-type` flag. Custom headers override default request headers such as `Accept` by name.

## Consequences

Agents can reach endpoints that require custom `Accept`, tracing, versioning, or conditional request headers without leaving the CLI surface.

The auth token and body media type each continue to have one owner, which keeps the raw escape hatch low-level without making it ambiguous.

## Rejected Alternatives

- Allowing arbitrary `Authorization` headers: complete at the HTTP layer, but it splits credential ownership and makes accidental token leakage easier.
- Using `--header Content-Type: ...` instead of `--content-type`: familiar for curl users, but it creates two ways to express the same body invariant.
