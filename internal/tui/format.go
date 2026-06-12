package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dutchcaz/youtrack/internal/youtrack"
)

func truncate(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	if width <= 1 {
		return value[:width]
	}
	if width <= 3 {
		return value[:width]
	}
	return value[:width-3] + "..."
}

func truncateBlock(value string, width, height int) string {
	lines := strings.Split(value, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i := range lines {
		lines[i] = truncate(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func trimBlank(value string) string {
	if strings.TrimSpace(value) == "" {
		return "No description"
	}
	return value
}

func resolvedText(value any) string {
	if value == nil {
		return "no"
	}
	return "yes"
}

func issueFields(issue youtrack.Issue) []string {
	if len(issue.Custom) == 0 {
		return nil
	}
	var fields []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(issue.Custom, &fields); err != nil {
		return nil
	}

	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		value := fieldValue(field.Value)
		if value == "" {
			continue
		}
		lines = append(lines, field.Name+": "+value)
	}
	return lines
}

func fieldValue(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return text
	}

	var object struct {
		Presentation string `json:"presentation"`
		Name         string `json:"name"`
		Login        string `json:"login"`
		Text         string `json:"text"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		return firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text)
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
			value := firstNonEmpty(object.Presentation, object.Name, object.Login, object.Text)
			if value != "" {
				values = append(values, value)
			}
		}
		return strings.Join(values, ", ")
	}

	var primitive any
	if err := json.Unmarshal(data, &primitive); err == nil {
		return fmt.Sprint(primitive)
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func nonEmpty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func formatBytes(size int64) string {
	if size <= 0 {
		return ""
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	value := float64(size)
	for _, suffix := range []string{"KiB", "MiB", "GiB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f TiB", value/unit)
}

func linkedIssues(link youtrack.IssueLink) []youtrack.Issue {
	if len(link.Trimmed) > 0 {
		return link.Trimmed
	}
	return link.Issues
}

func linkDirection(link youtrack.IssueLink) string {
	switch link.Direction {
	case "OUTWARD":
		if link.LinkType.SourceToTarget != "" {
			return link.LinkType.SourceToTarget
		}
	case "INWARD":
		if link.LinkType.TargetToSource != "" {
			return link.LinkType.TargetToSource
		}
	case "BOTH":
		if link.LinkType.SourceToTarget != "" {
			return link.LinkType.SourceToTarget
		}
	}
	return link.Direction
}
