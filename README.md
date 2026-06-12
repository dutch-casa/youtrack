# YouTrack CLI

Native Go CLI for JetBrains YouTrack. The default mode is non-interactive and JSON-first so agents can call it reliably. Interactive mode is available through Bubble Tea.

## Install

```sh
go install ./cmd/yt
```

## Authenticate

Create a YouTrack permanent token, then run:

```sh
yt auth login --url https://example.youtrack.cloud --token perm:...
```

You can also avoid local config with environment variables:

```sh
export YOUTRACK_URL=https://example.youtrack.cloud
export YOUTRACK_TOKEN=perm:...
```

If a command needs credentials and stdin is a terminal, `yt` prompts and saves them once. In non-interactive contexts, missing credentials return an error that names the setup command and env vars.

## Agent-Friendly Commands

```sh
yt me
yt projects list
yt users list
yt issues list --query 'project: ABC #Unresolved' --top 20
yt issues show ABC-123
yt issues create --project ABC --summary 'Fix login redirect' --description 'Observed in staging'
yt issues create --project ABC --summary 'Long bug report' --description-file ./report.md
yt issues update ABC-123 --summary 'Fix login redirect after SSO'
yt comments list ABC-123
yt comments add ABC-123 --text 'I can reproduce this.'
yt comments add ABC-123 --text-stdin < ./notes.md
yt work-items list ABC-123
yt work-items add ABC-123 --minutes 45 --text 'implementation'
yt commands apply ABC-123 --query 'State Fixed' --comment 'Fixed in main'
yt raw /api/admin/projects
```

JSON is the default. Use `--format table` for human-readable output.

`yt commands apply` is the high-leverage bridge to YouTrack's own command language. Use it for issue operations that humans normally perform through command input in the UI, including assignment, state changes, tags, links, watchers, and similar workflow actions, subject to the token's permissions.

For long generated text, description and comment commands accept explicit file/stdin sources such as `--description-file`, `--description-stdin`, `--text-file`, `--text-stdin`, `--comment-file`, and `--comment-stdin`.

## Interactive Mode

```sh
yt interactive --query 'project: ABC #Unresolved'
```

Keys: `j/k` move, `g/G` top/bottom, `r` refresh, `q` quit.
