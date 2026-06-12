# 0029 Raw Query Parameters

## Status

Accepted.

## Context

`yt raw` is the agent escape hatch for YouTrack REST capabilities that do not deserve a first-class command yet. Many YouTrack endpoints rely heavily on query parameters, especially `fields`, paging, filtering, and endpoint-specific switches.

Before this decision, callers could put query strings directly in `PATH`, but that forced agents to assemble and shell-quote complete URLs. It also made repeated parameters awkward and kept query handling as string manipulation at the CLI boundary.

## Decision

Add repeatable `yt raw --query name=value` / `-q name=value` flags and carry them through `youtrack.RawRequest.Query`.

The YouTrack client merges typed raw query values with any query string already present in `RawRequest.Path`. Repeated query names are preserved.

Keep this under `raw` instead of adding first-class command flags for endpoint-specific cases.

## Consequences

Agents can call query-rich YouTrack endpoints with named parameters while preserving the non-interactive command contract.

The raw client surface remains intent-shaped: callers provide a path plus query values, and the YouTrack client owns URL encoding and merge behavior.

## Rejected Alternatives

- Require callers to embed every query string in `PATH`: minimal code, but pushes URL encoding and repeated-parameter handling onto agents.
- Add first-class flags for common raw parameters such as `fields`: easier for one endpoint family, but it widens the raw command with YouTrack-specific speculation.
