# 0037 Issue Evidence and Article Heading Panels

## Status

Accepted

## Context

The issue detail pane now has a signal panel for fields, but related evidence still required tabbing through comments, links, activity, work items, and attachments with no summary in the default issue view. That made a selected issue feel under-instrumented until the user visited each pane.

Knowledge-base article detail also needs to preserve the article title as the first visible document signal. Article code, project, and author are useful metadata, but they should not visually displace the title. The article code still needs to remain visible at the top of the content pane because the left list may be scrolled away or visually separated from the body.

## Decision

Add an `Evidence` panel to the issue detail pane. It summarizes cached or loading state for comments, links, activity, work items, and attachments:

- Loaded panes show counts.
- Loading panes show `loading`.
- Failed panes show `error`.
- Unloaded panes show `tab to load` instead of triggering extra network requests on selection.

Keep the full evidence content in the existing tab panes.

Render knowledge-base article bodies code-and-title-first: the article ID/readable code and summary are the Markdown heading, while article ID, project, and author remain secondary metadata below it.

## Consequences

Clicking or selecting an issue now gives a compact evidence map without making issue selection fan out into five network calls. Users can decide which pane to load from the summary.

Knowledge-base articles keep both article identity and title visible in the scroller, while metadata remains available and can wrap or truncate without becoming the primary signal.

## Rejected

- Eager-load all issue evidence on every selection: richer counts, but too much latency and API traffic for list browsing.
- Hide unloaded evidence until visited: visually cleaner, but gives no indication of what the issue detail workspace can inspect.
- Put article project/author before the title: useful metadata, but it made the scroller feel like a record dump instead of a document.
