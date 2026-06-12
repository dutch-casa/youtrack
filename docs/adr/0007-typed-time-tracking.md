# 0007 Typed Time Tracking

## Status

Accepted

## Context

Time tracking is a common YouTrack workflow. It can be reached through raw REST and sometimes through YouTrack commands, but adding time has enough structure that a typed command reduces misuse:

- an issue is required
- duration is required
- duration must be positive
- optional text can come from literal, file, or stdin sources
- optional work item type, author, date, and notification muting should be named

## Decision

Expose work items as a first-class noun group:

```sh
yt work-items list ABC-123
yt work-items add ABC-123 --minutes 45 --text 'implementation'
```

Keep the REST details in `internal/youtrack` and the public command contract in `internal/ytcli`.

## Consequences

Agents can add and inspect time entries without hand-crafting JSON. The command rejects non-positive durations before authentication or network calls.

This does not replace `yt commands apply` or `yt raw`; it promotes one common workflow to a safer path because the required fields and validation are stable enough to encode.

## Rejected

- Only using `yt raw`: complete but too easy to send malformed duration bodies.
- Encoding duration presentation strings as the primary input: convenient for humans, less precise for agents. `--minutes` is explicit and stable.
- Hiding time tracking under `issues`: work items are their own YouTrack resource and deserve a noun group.
