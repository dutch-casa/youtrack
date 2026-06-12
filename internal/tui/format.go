package tui

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/dutch-casa/youtrack/internal/youtrack"
)

func truncate(value string, width int) string {
	value = terminalText(value)
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
	value = terminalText(value)
	if strings.TrimSpace(value) == "" {
		return "No description"
	}
	return value
}

func terminalText(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = strings.ReplaceAll(value, "\u202f", " ")
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.ReplaceAll(value, "\r", "\n")
}

func inlineText(value string) string {
	return strings.TrimSpace(terminalText(value))
}

func resolvedText(value any) string {
	if value == nil {
		return "no"
	}
	return "yes"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = inlineText(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func nonEmpty(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = inlineText(value)
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

func workItemLine(item youtrack.WorkItem) string {
	duration := firstNonEmpty(item.Duration.Presentation, formatMinutes(item.Duration.Minutes))
	author := firstNonEmpty(item.Author.FullName, item.Author.Name, item.Author.Login, item.Creator.FullName, item.Creator.Name, item.Creator.Login)
	kind := item.Type.Name
	date := formatMillisDate(item.Date)
	return strings.TrimSpace(strings.Join(nonEmpty(duration, kind, author, date), "  "))
}

func formatMinutes(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	hours := minutes / 60
	remaining := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", remaining)
	}
	if remaining == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, remaining)
}

func formatMillisDate(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.UnixMilli(value).UTC().Format("2006-01-02")
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
