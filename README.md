# YouTrack CLI

Native Go CLI for JetBrains YouTrack. The default mode is non-interactive and JSON-first so agents can call it reliably. Interactive mode is available through Bubble Tea.

## Install

Paste this into a terminal on a machine where Go is installed:

```sh
curl -fsSL https://raw.githubusercontent.com/dutch-casa/youtrack/main/scripts/install.sh | sh
```

That installs `yt` to `~/.local/bin` and registers `youtrack` as an alias for the same CLI.

To update later:

```sh
yt upgrade
```

If you already have the checkout:

```sh
scripts/install.sh
```

By default this installs `yt` and `youtrack` to `~/.local/bin`. To choose another location:

```sh
scripts/install.sh --bin-dir /usr/local/bin
scripts/install.sh --prefix "$HOME/.local"
scripts/install.sh --dry-run
```

## Authenticate

Authentication uses a YouTrack permanent token, not an OAuth browser grant. Create a token in YouTrack from Profile -> Account Security -> Tokens -> New token, then run:

```sh
yt auth login --url https://example.youtrack.cloud --token perm:...
```

`yt auth login` verifies the token with YouTrack before saving it. For offline/manual config, add `--no-verify`.

For the easiest guided setup, let `yt` open your YouTrack account security page and walk you through token creation:

```sh
yt auth login --open
```

This opens `<baseurl>/users/me?tab=account-security` and works for YouTrack Cloud and self-hosted instances because it preserves the instance base path and only needs a token generated in your own profile.

You can also avoid local config with environment variables:

```sh
export YOUTRACK_URL=https://example.youtrack.cloud
export YOUTRACK_TOKEN=perm:...
```

If a command needs credentials and stdin is a terminal, `yt` prompts, verifies the token, and saves it once. In non-interactive contexts, missing credentials return an error that names the setup command and env vars.

## Agent-Friendly Commands

```sh
yt capabilities
yt me
yt projects list
yt users list
yt articles list
yt articles list --project ABC
yt articles show ABC-A-1
yt agiles list
yt agiles sprints 120-1
yt helpdesk projects
yt helpdesk tickets SUPPORT --query '#Unresolved'
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
yt attachments list ABC-123
yt attachments add ABC-123 --file ./screenshot.png
yt activities list ABC-123
yt history list ABC-123 --category CommentsCategory --category CustomFieldCategory
yt links list ABC-123
yt commands apply ABC-123 --query 'State Fixed' --comment 'Fixed in main'
yt raw /api/admin/projects
yt raw /api/issues --method POST --body-file ./issue.json
yt raw /api/custom --method PATCH --content-type text/plain --body 'plain text'
yt raw /api/custom --header 'Accept: application/xml' --header 'X-YouTrack-Trace: agent-run-1'
yt raw /api/issues --query 'fields=id,idReadable,summary' --query '$top=10'
yt raw /api/files/123 --output-file ./download.bin
```

JSON is the default. Use `--format table` for human-readable output.

Use `yt capabilities` or `youtrack capabilities` as the machine-readable contract for agents. It runs without authentication and describes the default mode, supported command names, first-class commands, the interactive TUI, and the completeness bridges for YouTrack command-language workflows and raw REST calls.

`yt commands apply` is the high-leverage bridge to YouTrack's own command language. Use it for issue operations that humans normally perform through command input in the UI, including assignment, state changes, tags, links, watchers, and similar workflow actions, subject to the token's permissions.

`yt helpdesk` is intentionally issue-backed. It lists projects whose project type is Helpdesk and lists tickets through the public issue search surface with a `project:` query. For Helpdesk-specific channel or workflow endpoints that are not exposed as typed commands, use `yt raw` against the documented REST endpoint or an app-provided custom endpoint.

For long generated text, description and comment commands accept explicit file/stdin sources such as `--description-file`, `--description-stdin`, `--text-file`, `--text-stdin`, `--comment-file`, and `--comment-stdin`.

For long raw REST payloads, use `--body-file` or `--body-stdin` instead of shell-escaping large bodies. Raw requests default body content to `application/json`; use `--content-type` when a long-tail endpoint expects a different media type. Use repeated `--header` or `-H` flags for endpoint-specific headers and repeated `--query` or `-q` flags for query parameters. Raw stdout is byte-exact and does not add a newline. Use `--output-file` or `-o` for downloads or binary responses. `Authorization` remains managed by auth, and `Content-Type` remains managed by `--content-type`.

## Interactive Mode

```sh
yt interactive --query 'project: ABC #Unresolved'
```

Image attachments render inline when the terminal supports a known image protocol. Auto-detection covers Kitty, Ghostty, iTerm2, and compatible terminals; force a protocol with `yt interactive --image-protocol kitty` or `yt interactive --image-protocol iterm2`. Use `YOUTRACK_IMAGE_PROTOCOL=kitty` or `YT_IMAGE_PROTOCOL=kitty` to make that preference sticky.

Keys: `1-6` switch Issues, Knowledge Base, Help Desk, Agile Boards, Projects, and Users. Mouse clicks select sections and rows. The wheel moves the list on the left and scrolls content on the right. `j/k` move, `/` live-filters the current non-issue section or edits the issue query, `P` opens a project selector for issue filtering, `i` jumps to a specific issue ID, `o` opens the selected item in your browser, `c` adds a quick comment on issues, `w` adds a quick work item such as `45m implementation`, `n/p` moves between issue result pages, `tab` switches issue details/comments/links/activity/work/attachments, `pgup/pgdn` scrolls the selected pane, `:` applies a YouTrack command to the selected issue, `g/G` top/bottom, `r` refresh, `q` quit.
