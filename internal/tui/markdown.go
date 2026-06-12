package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
)

var youtrackImagePattern = regexp.MustCompile(`(?m)(^|[\s(])((?:[./]|https?://)[^\s)]+\.(?:png|jpe?g|gif|webp|bmp|svg))(?:\{[^}\n]*\})`)
var markdownImagePattern = regexp.MustCompile(`!\[([^\]\n]*)\]\(([^\s)\n]+)(?:\s+"[^"\n]*")?\)`)

type markdownImageRef struct {
	Alt string
	URL string
}

func renderMarkdown(value string, width int) string {
	value = markdownSource(value)
	if value == "" {
		return "No description"
	}
	return renderMarkdownSource(value, width)
}

func renderMarkdownSource(value string, width int) string {
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

func renderMarkdownWithImages(value string, width int, previews map[string]attachmentPreview, protocol imageProtocol) string {
	source := markdownSource(value)
	refs := markdownImageRefs(source)
	if len(refs) == 0 || protocol == imageProtocolNone {
		return renderMarkdownSource(source, width)
	}
	content := strings.TrimSpace(markdownImagePattern.ReplaceAllString(source, ""))
	rendered := renderMarkdownSource(content, width)
	lines := []string{rendered, "", titleStyle.Render("Images")}
	for _, ref := range refs {
		lines = append(lines, inlineText(firstNonEmpty(ref.Alt, ref.URL)))
		preview, ok := previews[ref.URL]
		switch {
		case !ok:
			lines = append(lines, helpStyle.Render("preview loading"))
		case preview.Err != nil:
			lines = append(lines, errorStyle.Render("preview unavailable: "+preview.Err.Error()))
		case len(preview.Data) > 0:
			imageWidth, imageHeight := inlineImageSize(width)
			lines = append(lines, renderInlineImage(protocol, inlineImage{
				Name:   firstNonEmpty(ref.Alt, ref.URL),
				Data:   preview.Data,
				Width:  imageWidth,
				Height: imageHeight,
			}))
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func markdownSource(value string) string {
	return strings.TrimSpace(normalizeYouTrackMarkdown(terminalText(value)))
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

func markdownImageRefs(value string) []markdownImageRef {
	matches := markdownImagePattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return nil
	}
	refs := make([]markdownImageRef, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		rawURL := strings.TrimSpace(match[2])
		if rawURL == "" || !isTerminalImageURL(rawURL) {
			continue
		}
		if _, ok := seen[rawURL]; ok {
			continue
		}
		seen[rawURL] = struct{}{}
		refs = append(refs, markdownImageRef{Alt: strings.TrimSpace(match[1]), URL: rawURL})
	}
	return refs
}

func isTerminalImageURL(rawURL string) bool {
	lower := strings.ToLower(strings.TrimSpace(rawURL))
	for _, extension := range []string{".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg"} {
		if strings.HasSuffix(lower, extension) || strings.Contains(lower, extension+"?") || strings.Contains(lower, extension+"#") {
			return true
		}
	}
	return false
}
