package textinput

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLiteral(t *testing.T) {
	text, ok, err := Resolve(Source{Name: "comment", Literal: "hello", LiteralSet: true}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !ok || text != "hello" {
		t.Fatalf("Resolve() = %q, %v", text, ok)
	}
}

func TestResolveEmptyLiteral(t *testing.T) {
	text, ok, err := Resolve(Source{Name: "description", Literal: "", LiteralSet: true}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !ok || text != "" {
		t.Fatalf("Resolve() = %q, %v", text, ok)
	}
}

func TestResolveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "body.txt")
	if err := os.WriteFile(path, []byte("from file"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	text, ok, err := Resolve(Source{Name: "description", File: path}, strings.NewReader(""))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !ok || text != "from file" {
		t.Fatalf("Resolve() = %q, %v", text, ok)
	}
}

func TestResolveStdin(t *testing.T) {
	text, ok, err := Resolve(Source{Name: "comment", Stdin: true}, strings.NewReader("from stdin"))
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !ok || text != "from stdin" {
		t.Fatalf("Resolve() = %q, %v", text, ok)
	}
}

func TestResolveRejectsMultipleSources(t *testing.T) {
	_, _, err := Resolve(Source{Name: "comment", Literal: "hello", LiteralSet: true, Stdin: true}, strings.NewReader(""))
	if err == nil {
		t.Fatal("Resolve() error = nil, want multiple source error")
	}
}
