package ytcli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpIsAvailableWithoutAuth(t *testing.T) {
	var out bytes.Buffer
	err := Execute(context.Background(), []string{"--help"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "Agent-friendly YouTrack CLI") {
		t.Fatalf("help output = %q", out.String())
	}
	for _, command := range []string{"projects", "users", "commands", "attachments", "activities", "links"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help output missing %q: %q", command, out.String())
		}
	}
	if !strings.Contains(out.String(), "work-items") {
		t.Fatalf("help output missing work-items: %q", out.String())
	}
}

func TestAuthLoginStatusLogout(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	var out bytes.Buffer

	err := Execute(context.Background(), []string{
		"--config", config,
		"auth", "login",
		"--url", "https://example.youtrack.cloud",
		"--token", "perm:secret",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth login error = %v", err)
	}

	out.Reset()
	err = Execute(context.Background(), []string{"--config", config, "auth", "status"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth status error = %v", err)
	}
	if strings.Contains(out.String(), "perm:secret") {
		t.Fatalf("auth status leaked token: %q", out.String())
	}
	if !strings.Contains(out.String(), "configured") {
		t.Fatalf("auth status output = %q", out.String())
	}

	out.Reset()
	err = Execute(context.Background(), []string{"--config", config, "auth", "logout"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("auth logout error = %v", err)
	}
}

func TestMissingAuthNonInteractiveReturnsSetupError(t *testing.T) {
	config := filepath.Join(t.TempDir(), "missing.json")
	err := Execute(context.Background(), []string{"--config", config, "me"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want missing auth")
	}
	if !strings.Contains(err.Error(), "yt auth login") {
		t.Fatalf("Execute() error = %q, want setup command", err.Error())
	}
}

func TestIssuesUpdateRequiresChangedField(t *testing.T) {
	config := filepath.Join(t.TempDir(), "config.json")
	err := Execute(context.Background(), []string{"--config", config, "issues", "update", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "--summary") {
		t.Fatalf("Execute() error = %q, want field guidance", err.Error())
	}
}

func TestCommentAddReadsTextFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/comments" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"c-1","text":"from stdin","author":{"login":"jane"},"created":1}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"comments", "add", "ABC-1", "--text-stdin",
	}, strings.NewReader("from stdin"), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "from stdin") {
		t.Fatalf("output = %q, want stdin comment", out.String())
	}
}

func TestWorkItemsAddRequiresPositiveMinutes(t *testing.T) {
	err := Execute(context.Background(), []string{"work-items", "add", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want minutes validation")
	}
	if !strings.Contains(err.Error(), "--minutes") {
		t.Fatalf("Execute() error = %q, want minutes guidance", err.Error())
	}
}

func TestAttachmentsAddRequiresFile(t *testing.T) {
	err := Execute(context.Background(), []string{"attachments", "add", "ABC-1"}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want file validation")
	}
	if !strings.Contains(err.Error(), "--file") {
		t.Fatalf("Execute() error = %q, want file guidance", err.Error())
	}
}

func TestAttachmentsAddUploadsFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/attachments" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart form", r.Header.Get("Content-Type"))
		}
		file, header, err := r.FormFile("upload")
		if err != nil {
			t.Fatalf("upload field: %v", err)
		}
		defer file.Close()
		if header.Filename != "evidence.txt" {
			t.Fatalf("filename = %q", header.Filename)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(file); err != nil {
			t.Fatalf("read upload: %v", err)
		}
		if body.String() != "evidence" {
			t.Fatalf("upload body = %q", body.String())
		}
		_, _ = w.Write([]byte(`[{"id":"134-1","name":"evidence.txt","size":8}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	path := filepath.Join(t.TempDir(), "evidence.txt")
	if err := os.WriteFile(path, []byte("evidence"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"attachments", "add", "ABC-1", "--file", path,
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "evidence.txt") {
		t.Fatalf("output = %q, want attachment name", out.String())
	}
}

func TestActivitiesListUsesDefaultCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/activities" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		categories := r.URL.Query().Get("categories")
		if !strings.Contains(categories, "CommentsCategory") || !strings.Contains(categories, "CustomFieldCategory") {
			t.Fatalf("categories = %q, want default activity categories", categories)
		}
		if got := r.URL.Query().Get("reverse"); got != "true" {
			t.Fatalf("reverse = %q, want true", got)
		}
		_, _ = w.Write([]byte(`[{"id":"a-1","$type":"CommentActivityItem","timestamp":1,"author":{"login":"jane"},"target":{"text":"Looks fixed"}}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"history", "list", "ABC-1",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "Looks fixed") {
		t.Fatalf("output = %q, want activity target", out.String())
	}
}

func TestActivitiesListNarrowsCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("categories"); got != "CommentsCategory" {
			t.Fatalf("categories = %q, want explicit category only", got)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"activities", "list", "ABC-1", "--category", "CommentsCategory",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestLinksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/links" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "5" {
			t.Fatalf("$top = %q, want 5", got)
		}
		_, _ = w.Write([]byte(`[{"id":"80-1s","direction":"OUTWARD","linkType":{"name":"Depend","sourceToTarget":"is required for","targetToSource":"depends on","directed":true},"issues":[{"id":"2-44","idReadable":"ABC-2","summary":"Linked issue"}]}]`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"--format", "table",
		"links", "list", "ABC-1", "--top", "5",
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "ABC-2") || !strings.Contains(out.String(), "is required for") {
		t.Fatalf("output = %q, want linked issue and direction", out.String())
	}
}
