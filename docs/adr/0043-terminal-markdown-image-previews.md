# 0043 Terminal Markdown Image Previews

## Status

Accepted

## Context

Knowledge-base articles can contain YouTrack image markup such as `/image.png{width=70%}`. Rendering that as formatted Markdown removes the raw sizing suffix, but it still leaves terminals without the image itself.

The TUI already supports issue attachment previews. Markdown images need the same terminal image protocols, but they are discovered from document text rather than from typed attachment metadata.

## Decision

Add a generic authenticated `FileContent` method in `internal/youtrack`. It reuses the same same-origin and base-path checks as attachment previews and keeps bearer-token byte downloads owned by the REST client boundary.

Keep Markdown image behavior in `internal/tui`:

- Normalize YouTrack image-size markup into Markdown image links.
- Extract terminal-image URLs from selected Markdown resources.
- Load only the selected resource's image previews through Bubble Tea commands.
- Cache previews by resource ID and image URL.
- Render images as capped thumbnails so they do not dominate the split-pane layout.
- Reserve terminal rows after each image escape sequence so later pane content does not slide under the rendered image.
- Clear Kitty terminal image placements before each image-capable frame so stale graphics do not remain after navigation.

## Consequences

Knowledge-base images can render inline in terminals that support Kitty or iTerm2-compatible image protocols.

The TUI does not prefetch every article image on a page, and `View` remains pure rendering over cached state.

The REST client remains responsible for URL safety. The TUI never sends bearer tokens by constructing its own HTTP requests.

## Rejected

- Leave Markdown image links as text: keeps the terminal unlike YouTrack for image-heavy articles.
- Let the TUI download arbitrary image URLs: leaks auth and URL-safety knowledge into terminal rendering.
- Prefetch all resource images: simple, but it wastes requests and can make article lists feel slow.
