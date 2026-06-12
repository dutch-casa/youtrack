package tui

import (
	"fmt"
	"strings"

	"github.com/dutch-casa/youtrack/internal/youtrack"
)

func issueFieldValue(issue youtrack.Issue, names ...string) string {
	return issue.CustomFieldValue(names...)
}

func issueFields(issue youtrack.Issue) []string {
	lines := make([]string, 0, len(issue.CustomFields))
	for _, field := range issue.CustomFields {
		name := strings.TrimSpace(field.Name)
		value := strings.TrimSpace(field.Value)
		if name == "" || value == "" {
			continue
		}
		lines = append(lines, name+": "+value)
	}
	return lines
}

func issueMetadataLine(issue youtrack.Issue) string {
	parts := make([]string, 0, 6)
	if project := firstNonEmpty(issue.Project.ShortName, issue.Project.Name); project != "" {
		parts = append(parts, "Project "+project)
	}
	for _, field := range []struct {
		label string
		names []string
	}{
		{label: "State", names: []string{"State"}},
		{label: "Assignee", names: []string{"Assignee", "Assignees"}},
		{label: "Priority", names: []string{"Priority"}},
		{label: "Type", names: []string{"Type"}},
	} {
		if value := issue.CustomFieldValue(field.names...); value != "" {
			parts = append(parts, field.label+" "+value)
		}
	}
	parts = append(parts, "Resolved "+resolvedText(issue.Resolved))
	return strings.Join(nonEmpty(parts...), "  ")
}

func issueListLine(issue youtrack.Issue) string {
	state := issueFieldValue(issue, "State")
	if state == "" {
		return fmt.Sprintf("%-12s %s", issue.IDReadable, issue.Summary)
	}
	return fmt.Sprintf("%-12s [%s] %s", issue.IDReadable, state, issue.Summary)
}
