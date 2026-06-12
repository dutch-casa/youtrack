# 0042 Issue-Backed Helpdesk Surface

## Status

Accepted

## Context

The CLI exposes `yt helpdesk projects` and `yt helpdesk tickets`. The current public YouTrack REST resource index exposes projects and issues, but does not present a separate first-class Helpdesk Tickets resource in the generated REST API reference.

Helpdesk tickets are represented in YouTrack as issues in Helpdesk projects for the read workflows this CLI supports. Helpdesk also has product-specific channel behavior, forms, email handling, customer visibility, and workflow semantics that should not be guessed from the regular issue API.

## Decision

Keep the typed `yt helpdesk` surface issue-backed:

- `yt helpdesk projects` lists projects and filters project type `helpdesk`.
- `yt helpdesk tickets PROJECT` lists issues with `project: PROJECT` plus the caller's additional query.

Document that this is a convenience over public project/issue resources, not a complete Helpdesk-specific API model.

For Helpdesk-specific or self-hosted app endpoints, agents should use `yt raw` until a stable public endpoint and common workflow justify a typed command.

## Consequences

The command remains useful for common support-agent browsing and filtering.

Agents do not infer unsupported ticket-channel creation or Helpdesk-only workflow semantics from the typed command name.

Future Helpdesk-specific commands should be added only when the API surface is explicit enough to model required fields and failure modes honestly.

## Rejected

- Remove `yt helpdesk`: loses a useful product-surface entry point for browsing Helpdesk projects and issue-backed tickets.
- Pretend regular issue commands cover all Helpdesk behavior: misleading for customer/channel-specific workflows.
- Reverse-engineer browser Helpdesk requests: fragile and outside the public REST contract.
