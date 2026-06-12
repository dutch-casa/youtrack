package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
)

var youtrackImagePattern = regexp.MustCompile(`(?m)(^|[\s(])((?:[./]|https?://)[^\s)]+\.(?:png|jpe?g|gif|webp|bmp|svg))(?:\{[^}\n]*\})`)

func renderMarkdown(value string, width int) string {
	value = strings.TrimSpace(normalizeYouTrackMarkdown(terminalText(value)))
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

func normalizeYouTrackMarkdown(value string) string {
	return youtrackImagePattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := youtrackImagePattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		prefix := parts[1]
		path := parts[2]
		name := path
		if slash := strings.LastIndex(name, "/"); slash >= 0 {
			name = name[slash+1:]
		}
		return prefix + "![" + name + "](" + path + ")"
	})
}
