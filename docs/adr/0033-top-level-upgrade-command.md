# 0033 Top-Level Upgrade Command

## Status

Accepted.

## Context

The public installer can be re-run to update the CLI, but users should not need to remember the install command or know where the binary was placed. The CLI has no other self-management operations.

## Decision

Expose updates as `yt upgrade`.

The command downloads the public installer script and runs it with `--bin-dir` and `--name` derived from the currently running executable. Users can override those values with flags, but the default updates the binary they invoked.

Do not add a `yt self update` namespace unless more self-management operations appear.

## Consequences

The update command is short, discoverable from root help, and aligned with the one maintenance job users came to do.

The implementation reuses the installer instead of duplicating build/install logic in Go. Installer behavior has one owner.

## Rejected Alternatives

- `yt self update`: accurate, but shallow while there is only one self-management operation.
- Document re-running the install command only: works, but hides an expected lifecycle action from the CLI surface.
- Use `go install github.com/dutch-casa/youtrack/cmd/yt@latest`: simple, but installs into Go's binary directory instead of reliably updating the binary the user invoked.
