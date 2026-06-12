# 0039 Dual Command Name Registration

## Status

Accepted

## Context

The short command name `yt` is fast for agents and terminal users. The full command name `youtrack` is easier to discover, easier to remember from the product name, and clearer in scripts that may be read by people who do not know the abbreviation.

Both names should invoke the same CLI contract. They should not drift into separate binaries or different behavior.

## Decision

The installer builds one binary and registers both names:

- installing `yt` also creates a `youtrack` alias
- installing `youtrack` also creates a `yt` alias

The alias is a symlink to the installed binary name. `yt upgrade` and `youtrack upgrade` both call the same installer path, so upgrading from either command refreshes the counterpart alias.

Expose both names in `yt capabilities` through `commandNames`.

## Consequences

Agents and humans can choose either `yt` or `youtrack` without changing command semantics.

The install script remains simple: one build, one primary install target, one counterpart alias.

Custom `--name` installs remain custom and do not invent extra aliases for arbitrary names.

## Rejected

- Build two binaries: duplicates work and creates unnecessary room for drift.
- Rename the primary command to only `youtrack`: more descriptive, but worse for repeated agent and terminal use.
- Keep only `yt`: compact, but less discoverable and less obvious in shared scripts.
