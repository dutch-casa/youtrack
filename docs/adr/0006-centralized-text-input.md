# 0006 Centralized Text Input

## Status

Accepted

## Context

Agents and humans often need to send multiline issue descriptions, comments, and command comments. A single string flag is convenient for short text but poor for generated reports, stack traces, markdown, and pasted diagnostics.

Adding file and stdin flags independently to every command would duplicate subtle behavior:

- how to reject multiple text sources
- how to distinguish omitted text from an explicitly empty value
- how errors name the text being read

## Decision

Put text-source resolution in `internal/textinput`.

Commands declare their own public flags, then pass a `textinput.Source` with the text name, literal value, literal presence, file path, and stdin choice. The resolver returns the resolved text and whether any source was provided.

## Consequences

Commands can support literal, file, and stdin text without repeating file-reading rules. The command layer remains responsible for deciding whether text is required for a particular operation.

Explicit stdin flags are required. The CLI does not infer stdin text implicitly because the default mode is agent-first and missing authentication must not accidentally consume an input stream intended for another tool.

## Rejected

- Implicitly reading stdin when a flag is missing: convenient in shells, risky for agent pipelines and auth prompting.
- Per-command file reading: easy to start, likely to drift in validation and error wording.
- Treating empty string as always omitted: issue updates need to be able to clear a description intentionally.
