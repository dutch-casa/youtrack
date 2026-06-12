package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteRejectsUnsupportedFormat(t *testing.T) {
	var out bytes.Buffer
	err := Write(&out, Format("yaml"), map[string]string{"ok": "true"})
	if err == nil {
		t.Fatal("Write() error = nil, want unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("Write() error = %q, want unsupported format", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty output on unsupported format", out.String())
	}
}
