package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

func renderMarkdown(value string, width int) string {
	value = strings.TrimSpace(terminalText(value))
	if value == "" {
		return "No description"
	}
	if width < 20 {
		width = 20
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return value
	}
	rendered, err := renderer.Render(value)
	if err != nil {
		return value
	}
	rendered = strings.TrimSpace(rendered)
	if rendered == "" {
		return value
	}
	return rendered
}
