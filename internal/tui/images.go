package tui

import (
	"encoding/base64"
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

type attachmentPreview struct {
	Data []byte
	Err  error
}

func detectImageProtocol() imageProtocol {
	switch {
	case os.Getenv("KITTY_WINDOW_ID") != "" || strings.Contains(os.Getenv("TERM"), "kitty"):
		return imageProtocolKitty
	case os.Getenv("TERM_PROGRAM") == "iTerm.app" || os.Getenv("TERM_PROGRAM") == "WezTerm" || os.Getenv("WEZTERM_EXECUTABLE") != "":
		return imageProtocolITerm
	default:
		return imageProtocolNone
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
	imageWidth := min(max(12, width-6), 80)
	imageHeight := max(6, imageWidth/3)
	return renderInlineImage(m.imageProtocol, inlineImage{
		Name:   attachment.Name,
		Data:   preview.Data,
		Width:  imageWidth,
		Height: imageHeight,
	})
}

type inlineImage struct {
	Name   string
	Data   []byte
	Width  int
	Height int
}

func renderInlineImage(protocol imageProtocol, image inlineImage) string {
	switch protocol {
	case imageProtocolITerm:
		return renderITermImage(image)
	case imageProtocolKitty:
		return renderKittyImage(image)
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
