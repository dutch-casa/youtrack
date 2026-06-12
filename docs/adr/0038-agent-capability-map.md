# 0038 Agent Capability Map

## Status

Accepted

## Context

Agents need to use the CLI without repeatedly scraping `--help`. Help text is for humans and Cobra formatting can change, while agents need a stable contract that names:

- the default non-interactive JSON behavior
- supported command names
- first-class typed commands
- completeness bridges for command-language workflows and raw REST
- the optional human TUI

This discovery must not require YouTrack credentials, otherwise agents cannot plan setup before authentication exists.

## Decision

Add `yt capabilities` as an auth-free JSON command owned by `internal/ytcli`.

The document uses a schema version and explicitly reports:

- `commandNames`, currently `yt` and `youtrack`
- `defaultMode: non-interactive`
- `defaultOutput: json`
- authentication setup commands
- first-class command groups
- `yt commands apply` as the issue workflow bridge
- `yt raw` as the complete REST bridge for endpoints permitted by the token
- `yt interactive` as an additive human browsing surface

Keep the capabilities document hand-authored instead of derived from Cobra command traversal. It is a product contract, not a mirror of implementation flags.

## Consequences

Agents can call `yt capabilities` once, cache the result, and choose the correct surface without parsing help output.

The command makes the completeness model explicit: typed commands are preferred for common stable work, `commands apply` covers YouTrack command-language issue workflows, and `raw` covers long-tail REST.

Future changes to this JSON shape require schema-version consideration because agents may depend on it.

## Rejected

- Rely on `yt --help`: good for humans, but too loose and formatting-oriented for agents.
- Generate capabilities from Cobra automatically: reduces duplication, but it cannot explain semantic coverage such as when to use `commands apply` versus `raw`.
- Require authentication: blocks setup and planning flows.
