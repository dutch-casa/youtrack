package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dutch-casa/youtrack/internal/youtrack"
)

type issueFieldRow struct {
	Label string
	Value string
	Kind  string
}

func issueFieldValue(issue youtrack.Issue, names ...string) string {
	return issue.CustomFieldValue(names...)
}

func issueFieldPanel(issue youtrack.Issue, width int) string {
	rows := issueFieldRows(issue)
	if len(rows) == 0 {
		return ""
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, titleStyle.Render("Signals"))
	for _, row := range rows {
		lines = append(lines, fieldLabelStyle.Render(row.Label+":")+" "+issueFieldStyle(row).Render(row.Value))
	}
	panelWidth := max(20, width-2)
	return fieldPanelStyle.Width(panelWidth).Render(strings.Join(lines, "\n"))
}

func issueFieldRows(issue youtrack.Issue) []issueFieldRow {
	rows := make([]issueFieldRow, 0, len(issue.CustomFields)+2)
	seen := make(map[string]bool)
	if project := firstNonEmpty(issue.Project.ShortName, issue.Project.Name); project != "" {
		rows = append(rows, issueFieldRow{Label: "Project", Value: project, Kind: "project"})
	}
	for _, signal := range []struct {
		label string
		kind  string
		names []string
	}{
		{label: "State", kind: "state", names: []string{"State"}},
		{label: "Assignee", kind: "person", names: []string{"Assignee", "Assignees"}},
		{label: "Priority", kind: "priority", names: []string{"Priority"}},
		{label: "Type", kind: "type", names: []string{"Type"}},
	} {
		if value := inlineText(issue.CustomFieldValue(signal.names...)); value != "" {
			rows = append(rows, issueFieldRow{Label: signal.label, Value: value, Kind: signal.kind})
			for _, name := range signal.names {
				seen[strings.ToLower(name)] = true
			}
		}
	}
	rows = append(rows, issueFieldRow{Label: "Resolved", Value: resolvedText(issue.Resolved), Kind: "resolved"})
	for _, field := range issue.CustomFields {
		name := inlineText(field.Name)
		value := inlineText(field.Value)
		if name == "" || value == "" {
			continue
		}
		if seen[strings.ToLower(name)] {
			continue
		}
		rows = append(rows, issueFieldRow{Label: name, Value: value, Kind: "custom"})
	}
	return rows
}

func issueFieldStyle(row issueFieldRow) lipgloss.Style {
	value := strings.ToLower(row.Value)
	switch row.Kind {
	case "project":
		return fieldProjectStyle
	case "person":
		return fieldPersonStyle
	case "type":
		return fieldInfoStyle
	case "resolved":
		if value == "yes" {
			return fieldGoodStyle
		}
		return fieldWarnStyle
	case "state":
		if containsAny(value, "fixed", "done", "resolved", "closed", "verified") {
			return fieldGoodStyle
		}
		if containsAny(value, "blocked", "critical", "failed", "reopened") {
			return fieldBadStyle
		}
		if containsAny(value, "progress", "review", "testing") {
			return fieldInfoStyle
		}
		return fieldWarnStyle
	case "priority":
		if containsAny(value, "critical", "blocker", "show-stopper", "major", "high") {
			return fieldBadStyle
		}
		if containsAny(value, "normal", "medium") {
			return fieldWarnStyle
		}
		return fieldGoodStyle
	default:
		return fieldDefaultStyle
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func issueListLine(issue youtrack.Issue) string {
	state := issueFieldValue(issue, "State")
	if state == "" {
		return fmt.Sprintf("%-12s %s", issue.IDReadable, inlineText(issue.Summary))
	}
	return fmt.Sprintf("%-12s [%s] %s", issue.IDReadable, state, inlineText(issue.Summary))
}
