# 0041 Terminal Image Attachment Previews

## Status

Accepted

## Context

The TUI can browse issue attachments, but image attachments were listed only as file metadata. Terminals that support inline image protocols can show screenshots and other visual evidence directly in the issue workspace.

The preview path crosses two different bodies of knowledge:

- `internal/youtrack` knows how authenticated attachment bytes are fetched from YouTrack.
- `internal/tui` knows whether the current terminal supports inline image escape protocols and how to render terminal output.

These should not be merged. The YouTrack client should not know terminal protocols, and the TUI should not build raw attachment HTTP requests.

## Decision

Add `Client.AttachmentContent` in `internal/youtrack`. It downloads attachment thumbnail bytes when available, falls back to attachment URL bytes, enforces same-origin/base-path safety, sends the bearer token only to the configured YouTrack origin, and caps preview downloads.

Add optional attachment previews in `internal/tui`:

- Detect Kitty and iTerm2-compatible terminal image support from environment variables.
- Load previews asynchronously through Bubble Tea commands after attachment metadata is loaded.
- Render cached previews below image attachment metadata.
- Keep unsupported terminals text-only.
- Keep attachment previews cached with the attachment list and clear them when attachment evidence is invalidated.

## Consequences

Screenshots and other image attachments can appear inline in terminals that support the image protocol.

Unsupported terminals keep the existing attachment list behavior.

Preview loading does not block `View`; it follows the existing Bubble Tea message/update flow.

The TUI does not send tokens to arbitrary URLs returned in attachment metadata.

## Rejected

- Render every attachment as an image attempt: wastes API calls and produces noisy failures for non-image files.
- Let the TUI call `yt raw` or construct download URLs itself: leaks YouTrack REST knowledge into terminal rendering.
- Require an image-capable terminal: violates the additive TUI contract.
- Download full-size images unbounded: risks slow panes and excessive memory use.
