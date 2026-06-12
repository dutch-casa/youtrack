# 0023 Raw Request Headers

## Status

Accepted.

## Context

`yt raw` is the completeness boundary for YouTrack REST operations that are not yet modeled as typed commands. Some long-tail endpoints require request headers beyond method, path, body, and content type.

The raw command still must not let callers bypass the CLI's credential model. Agents should be able to set endpoint-specific headers, but bearer authorization should remain owned by `internal/auth` and `internal/youtrack`.

## Decision

Add repeated `yt raw --header/-H 'Name: value'` flags and carry them through `youtrack.RawRequest.Headers`.

Keep raw header policy in `internal/youtrack`. The CLI parser only splits `Name: value` text; it delegates header-name validation, managed `Authorization` rejection, and managed `Content-Type` rejection to the YouTrack client boundary. Custom headers override default request headers such as `Accept` by name.

## Consequences

Agents can reach endpoints that require custom `Accept`, tracing, versioning, or conditional request headers without leaving the CLI surface.

The auth token and body media type each continue to have one owner, which keeps the raw escape hatch low-level without making it ambiguous.

The parser and client no longer duplicate raw header policy, so changes to managed headers or HTTP header-name rules land in one module.

## Rejected Alternatives

- Allowing arbitrary `Authorization` headers: complete at the HTTP layer, but it splits credential ownership and makes accidental token leakage easier.
- Using `--header Content-Type: ...` instead of `--content-type`: familiar for curl users, but it creates two ways to express the same body invariant.
