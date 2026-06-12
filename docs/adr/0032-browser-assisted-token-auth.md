# 0032 Browser-Assisted Token Auth

## Status

Accepted.

## Context

The CLI needs authentication that is easy for humans to set up once and reliable for agents to use afterward. It also needs to work across YouTrack Cloud and self-hosted YouTrack installations.

YouTrack supports permanent tokens for REST API clients. YouTrack OAuth is provided by Hub, but YouTrack Cloud and built-in Hub support implicit OAuth grants rather than a CLI-friendly authorization-code flow with a local callback and refresh token. The implicit grant returns the access token in the browser URL fragment, which a local HTTP callback cannot receive directly.

## Decision

Keep permanent-token authentication as the default and universal CLI auth path.

Add browser-assisted setup with `yt auth login --open`. The command opens `<baseurl>/users/me?tab=account-security`, prints the token creation path, prompts for the permanent token, verifies the token against `/api/users/me`, and saves it with the existing local credential store.

Credential verification is enabled by default for `yt auth login` and for the automatic prompt shown when another command needs credentials. `--no-verify` exists for explicit offline/manual login configuration, but the normal successful setup means the CLI has already proven it can authenticate to the instance.

Do not make OAuth the default setup path. A future OAuth command may be added as an advanced mode for installations that have a suitable Hub client registration and redirect page, but it must not replace permanent-token setup.

## Consequences

The simplest path works for Cloud, self-hosted built-in Hub, and external Hub installations without requiring administrator-created OAuth clients.

The browser opens the stable account-security tab for the current user while preserving self-hosted base paths such as `/youtrack`. The CLI does not depend on a token fragment handoff page.

Failed verification does not write credentials. This keeps one-time setup from storing a bad token that later fails in agent workflows.

Prompt parsing must support terminal and piped input. The combined URL/token prompt reads from one buffered stream so non-terminal setup can pass both values without losing buffered data between prompts.

Agents retain the same non-interactive credential contract: saved config or `YOUTRACK_URL` and `YOUTRACK_TOKEN`.

## Rejected Alternatives

- Default to OAuth: more familiar in some CLIs, but YouTrack's commonly available OAuth surface is implicit grant, lacks refresh tokens, and is awkward for native CLIs.
- Deep-link to a private token creation endpoint beyond the account-security tab: faster when it works, but brittle across Cloud, self-hosted, and YouTrack version changes.
- Generate permanent tokens through the Hub API: not a self-service default for normal users and risks requiring broader administration permissions than the CLI needs.
