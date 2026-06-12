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
yt issues list --query 'project: ABC #Unresolved' --top 20
yt issues show ABC-123
yt issues create --project ABC --summary 'Fix login redirect' --description 'Observed in staging'
yt comments list ABC-123
yt comments add ABC-123 --text 'I can reproduce this.'
yt raw /api/admin/projects
```

JSON is the default. Use `--format table` for human-readable output.

## Interactive Mode

```sh
yt interactive --query 'project: ABC #Unresolved'
```

Keys: `j/k` move, `g/G` top/bottom, `r` refresh, `q` quit.
