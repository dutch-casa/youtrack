# 0031 Custom Field Projection Boundary

## Status

Accepted.

## Context

YouTrack issue custom fields arrive as endpoint-specific JSON values. A field value can be a string, a scalar, an object with `presentation`, `name`, `login`, or `text`, or a list of those objects.

The TUI needs simple display values for metadata strips and issue list state signals, but it should not know YouTrack's custom-field wire shape. Keeping that decoder in `internal/tui` made custom-field representation a Parnas failure: changing the API field selector or value shape required edits in both `internal/youtrack` and `internal/tui`.

## Decision

Move custom-field wire projection into `internal/youtrack`.

Expose custom fields as `Issue.CustomFields []CustomField`, where `CustomField.Value` is the display-ready value derived from YouTrack's wire representation. Add `Issue.CustomFieldValue` for case-insensitive lookup by field name.

Keep TUI-owned code responsible only for terminal presentation: choosing which field names to show, formatting metadata text, and skipping empty display rows.

## Consequences

YouTrack API custom-field changes now land in one module.

The TUI can still change its first-screen signals without learning REST JSON value variants.

## Rejected Alternatives

- Keep `json.RawMessage` on `Issue` and decode in each consumer: flexible, but it leaks YouTrack wire knowledge through the whole codebase.
- Add TUI-specific typed fields such as `State` and `Priority` to `Issue`: convenient for the current UI, but too narrow for project-specific YouTrack fields and future non-TUI consumers.
