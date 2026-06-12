package tui

import (
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/dutch-casa/youtrack/internal/youtrack"
)

const maxInlineImageBytes int64 = 4 << 20

type imageProtocol string

const (
	imageProtocolNone  imageProtocol = ""
	imageProtocolITerm imageProtocol = "iterm2"
	imageProtocolKitty imageProtocol = "kitty"
)

var errInvalidImageProtocol = errors.New("image protocol must be auto, none, kitty, or iterm2")

func ValidateImageProtocol(value string) error {
	_, err := parseImageProtocolName(value)
	return err
}

type attachmentPreview struct {
	Data []byte
	Err  error
}

func detectImageProtocol() imageProtocol {
	switch {
	case os.Getenv("KITTY_WINDOW_ID") != "" || strings.Contains(os.Getenv("TERM"), "kitty"):
		return imageProtocolKitty
	case strings.EqualFold(os.Getenv("TERM_PROGRAM"), "ghostty") || strings.Contains(os.Getenv("TERM"), "ghostty") || os.Getenv("GHOSTTY_RESOURCES_DIR") != "":
		return imageProtocolKitty
	case os.Getenv("TERM_PROGRAM") == "iTerm.app" || os.Getenv("TERM_PROGRAM") == "WezTerm" || os.Getenv("WEZTERM_EXECUTABLE") != "":
		return imageProtocolITerm
	default:
		return imageProtocolNone
	}
}

func resolveImageProtocol(value string) (imageProtocol, error) {
	value = requestedImageProtocol(value)
	name, err := parseImageProtocolName(value)
	if err != nil {
		return imageProtocolNone, err
	}
	switch name {
	case "", "auto":
		return detectImageProtocol(), nil
	case "none":
		return imageProtocolNone, nil
	case "kitty":
		return imageProtocolKitty, nil
	case "iterm2":
		return imageProtocolITerm, nil
	}
	return imageProtocolNone, nil
}

func requestedImageProtocol(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && !strings.EqualFold(value, "auto") {
		return value
	}
	if env := strings.TrimSpace(os.Getenv("YOUTRACK_IMAGE_PROTOCOL")); env != "" {
		return env
	}
	if env := strings.TrimSpace(os.Getenv("YT_IMAGE_PROTOCOL")); env != "" {
		return env
	}
	return value
}

func normalizeImageProtocolName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "iterm", "iterm2", "i-term2":
		return "iterm2"
	default:
		return value
	}
}

func parseImageProtocolName(value string) (string, error) {
	name := normalizeImageProtocolName(value)
	switch name {
	case "", "auto", "none", "kitty", "iterm2":
		return name, nil
	default:
		return "", errInvalidImageProtocol
	}
}

func isImageAttachment(attachment youtrack.Attachment) bool {
	mimeType := strings.ToLower(strings.TrimSpace(attachment.MimeType))
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	extension := strings.ToLower(strings.TrimPrefix(firstNonEmpty(attachment.Extension, filepath.Ext(attachment.Name)), "."))
	if extension == "" {
		return false
	}
	return strings.HasPrefix(mime.TypeByExtension("."+extension), "image/")
}

func attachmentPreviewID(attachment youtrack.Attachment) string {
	return firstNonEmpty(attachment.ID, attachment.Name)
}

func (m model) renderedAttachmentPreview(attachment youtrack.Attachment, width int) string {
	if m.imageProtocol == imageProtocolNone || !isImageAttachment(attachment) {
		return ""
	}
	issueID := m.currentIssueID()
	preview, ok := m.attachmentPreviews[issueID][attachmentPreviewID(attachment)]
	if !ok {
		return helpStyle.Render("preview loading")
	}
	if preview.Err != nil {
		return errorStyle.Render("preview unavailable: " + preview.Err.Error())
	}
	if len(preview.Data) == 0 {
		return ""
	}
	imageWidth, imageHeight := inlineImageSize(width)
	return renderInlineImage(m.imageProtocol, inlineImage{
		Name:   attachment.Name,
		Data:   preview.Data,
		Width:  imageWidth,
		Height: imageHeight,
	})
}

func inlineImageSize(width int) (int, int) {
	imageWidth := min(max(12, width-8), 48)
	imageHeight := min(max(4, imageWidth/4), 10)
	return imageWidth, imageHeight
}

type inlineImage struct {
	Name   string
	Data   []byte
	Width  int
	Height int
}

func renderInlineImage(protocol imageProtocol, image inlineImage) string {
	var rendered string
	switch protocol {
	case imageProtocolITerm:
		rendered = renderITermImage(image)
	case imageProtocolKitty:
		rendered = renderKittyImage(image)
	default:
		return ""
	}
	return rendered + strings.Repeat("\n", max(1, image.Height))
}

func clearInlineImages(protocol imageProtocol) string {
	switch protocol {
	case imageProtocolKitty:
		return "\x1b_Ga=d,d=A\x1b\\"
	default:
		return ""
	}
}

func renderITermImage(image inlineImage) string {
	name := base64.StdEncoding.EncodeToString([]byte(filepath.Base(image.Name)))
	data := base64.StdEncoding.EncodeToString(image.Data)
	return fmt.Sprintf("\x1b]1337;File=name=%s;inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\a", name, image.Width, image.Height, data)
}

func renderKittyImage(image inlineImage) string {
	data := base64.StdEncoding.EncodeToString(image.Data)
	var builder strings.Builder
	for len(data) > 0 {
		chunk := data
		if len(chunk) > 4096 {
			chunk = data[:4096]
		}
		data = data[len(chunk):]
		more := 0
		if len(data) > 0 {
			more = 1
		}
		if builder.Len() == 0 {
			fmt.Fprintf(&builder, "\x1b_Gf=100,t=d,a=T,s=%d,v=%d,m=%d;%s\x1b\\", image.Width, image.Height, more, chunk)
			continue
		}
		fmt.Fprintf(&builder, "\x1b_Gm=%d;%s\x1b\\", more, chunk)
	}
	return builder.String()
}
