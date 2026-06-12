package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dutchcaz/youtrack/internal/youtrack"
)

type issueCustomField struct {
	Name  string
	Value string
}

func issueCustomFields(issue youtrack.Issue) []issueCustomField {
	if len(issue.Custom) == 0 {
		return nil
	}

	var rawFields []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(issue.Custom, &rawFields); err != nil {
		return nil
	}

	fields := make([]issueCustomField, 0, len(rawFields))
	for _, rawField := range rawFields {
		name := strings.TrimSpace(rawField.Name)
		value := fieldValue(rawField.Value)
		if name == "" || value == "" {
			continue
		}
		fields = append(fields, issueCustomField{Name: name, Value: value})
	}
	return fields
}

func issueFieldValue(issue youtrack.Issue, names ...string) string {
	return customFieldValue(issueCustomFields(issue), names...)
}

func customFieldValue(fields []issueCustomField, names ...string) string {
	for _, field := range fields {
		for _, name := range names {
			if strings.EqualFold(field.Name, name) {
				return field.Value
			}
		}
	}
	return ""
}

func issueFields(issue youtrack.Issue) []string {
	fields := issueCustomFields(issue)
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		lines = append(lines, field.Name+": "+field.Value)
	}
	return lines
}

func issueMetadataLine(issue youtrack.Issue) string {
	fields := issueCustomFields(issue)
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
		if value := customFieldValue(fields, field.names...); value != "" {
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

func fieldValue(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return strings.TrimSpace(text)
	}

	var object struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		return strings.TrimSpace(firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text))
	}

	var objects []struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &objects); err == nil {
		values := make([]string, 0, len(objects))
		for _, object := range objects {
			value := strings.TrimSpace(firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text))
			if value != "" {
				values = append(values, value)
			}
		}
		return strings.Join(values, ", ")
	}

	var primitive any
	if err := json.Unmarshal(data, &primitive); err == nil {
		return strings.TrimSpace(fmt.Sprint(primitive))
	}
	return ""
}
