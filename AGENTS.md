# Agent Working Contract

This repository is an agent-first Go CLI for JetBrains YouTrack.

Before changing code:

- Load and follow the local `youtrack-cli`, `knowledge-decomposition`, and `interface-design` skills when available.
- For Go work, also use the local Go doctrine and relevant `software-*` and `test-*` skills.
- For TUI work, also use the local Charm TUI skill.

Design rules:

- Preserve the default non-interactive JSON CLI contract.
- Keep interactive TUI behavior additive; it must not become required for agent workflows.
- Put changes where the owning knowledge lives:
  - `internal/auth` owns credential discovery, prompting, and persistence.
  - `internal/youtrack` owns YouTrack REST wire knowledge.
  - `internal/ytcli` owns the command contract.
  - `internal/output` owns JSON/table representation.
  - `internal/tui` owns terminal interaction and rendering state.
- Prefer first-class typed commands for common stable workflows.
- Use `yt commands apply` for broad YouTrack command-language issue workflows.
- Preserve `yt raw` as the complete REST escape hatch for everything allowed by the token.
- Record non-obvious boundary or interface decisions as ADRs under `docs/adr/` and link them from `docs/module-guide.md`.

Commit rules:

- Commit atomically: one coherent behavioral or documentation decision per commit.
- Use the commit message as durable documentation of what changed.
- Do not credit tools, models, or assistants in commit messages.
- Do not mix unrelated refactors with feature or bug-fix commits.

Verification before a code commit:

```sh
gofmt -w .
go test ./...
go vet ./...
go run ./cmd/yt --help
```

For TUI changes, also run:

```sh
go run ./cmd/yt interactive --help
```

For installer or command-discovery changes, verify the relevant command directly, such as:

```sh
go run ./cmd/yt capabilities
sh scripts/install.sh --dry-run
```
