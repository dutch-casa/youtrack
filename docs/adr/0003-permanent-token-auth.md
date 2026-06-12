# 0003 Permanent-Token Authentication

## Status

Accepted

## Context

YouTrack supports permanent tokens for REST API clients. This CLI is a local native tool, not a browser-based app, so OAuth would add a multi-step authorization flow that does not improve the normal agent setup path.

## Decision

Use YouTrack permanent tokens with `Authorization: Bearer <token>`.

Credential sources are checked in this order:

1. `YOUTRACK_URL` and `YOUTRACK_TOKEN`
2. saved config file
3. terminal prompt, only when stdin is interactive

Saved credentials live under the user config directory with directory mode `0700` and file mode `0600`. Token values are redacted in status output.

## Consequences

One-time setup is simple:

```sh
yt auth login --url https://example.youtrack.cloud --token perm:...
```

Agents can avoid disk state entirely with environment variables. Missing credentials in non-interactive mode are a clear setup error, not a prompt.

## Rejected

- OAuth as the default: better for delegated browser flows, worse for local agent setup.
- Keychain integration as the only store: useful later, but not portable enough for the first cross-platform path.
- Prompting on every command: simple to implement, but bad for agent reliability.
