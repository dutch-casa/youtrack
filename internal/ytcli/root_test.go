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
	"testing/quick"
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

func TestRawReadsBodyFromFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != `{"summary":"from file"}` {
			t.Fatalf("body = %q", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	path := filepath.Join(t.TempDir(), "body.json")
	if err := os.WriteFile(path, []byte(`{"summary":"from file"}`), 0o600); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	var out bytes.Buffer
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--method", "POST", "--body-file", path,
	}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), `"ok":true`) {
		t.Fatalf("output = %q, want raw response", out.String())
	}
}

func TestRawReadsBodyFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != `{"summary":"from stdin"}` {
			t.Fatalf("body = %q", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "-X", "POST", "--body-stdin",
	}, strings.NewReader(`{"summary":"from stdin"}`), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsCustomContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "text/plain" {
			t.Fatalf("Content-Type = %q, want text/plain", got)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(r.Body); err != nil {
			t.Fatalf("read body: %v", err)
		}
		if body.String() != "plain text" {
			t.Fatalf("body = %q, want plain text", body.String())
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/custom", "-X", "POST", "--content-type", "text/plain", "--body", "plain text",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("Authorization = %q, want configured bearer token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/xml" {
			t.Fatalf("Accept = %q, want custom accept header", got)
		}
		if got := r.Header.Get("X-Youtrack-Trace"); got != "agent-run-1" {
			t.Fatalf("X-YouTrack-Trace = %q, want agent-run-1", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/custom", "-H", "Accept: application/xml", "--header", "X-YouTrack-Trace: agent-run-1",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawSendsQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query()["fields"]; len(got) != 1 || got[0] != "id,idReadable" {
			t.Fatalf("fields query = %#v, want id,idReadable", got)
		}
		if got := r.URL.Query()["tag"]; len(got) != 2 || got[0] != "agent" || got[1] != "urgent" {
			t.Fatalf("tag query = %#v, want repeated values", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("YOUTRACK_URL", server.URL)
	t.Setenv("YOUTRACK_TOKEN", "perm:test")

	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--query", "fields=id,idReadable", "-q", "tag=agent", "-q", "tag=urgent",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRawRejectsInvalidQueryBeforeAuth(t *testing.T) {
	err := Execute(context.Background(), []string{
		"--config", filepath.Join(t.TempDir(), "missing.json"),
		"raw", "/api/issues", "--query", "fields",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want query validation")
	}
	if !strings.Contains(err.Error(), "name=value") {
		t.Fatalf("Execute() error = %q, want query shape guidance", err.Error())
	}
	if strings.Contains(err.Error(), "yt auth login") {
		t.Fatalf("Execute() error = %q, parsed query after auth", err.Error())
	}
}

func TestRawRejectsManagedHeadersBeforeAuth(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "authorization",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "raw", "/api/issues", "-H", "Authorization: Bearer bad"},
			want: "managed by yt auth",
		},
		{
			name: "content type",
			args: []string{"--config", filepath.Join(t.TempDir(), "missing.json"), "raw", "/api/issues", "-H", "Content-Type: text/plain"},
			want: "use --content-type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Execute(context.Background(), tt.args, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil {
				t.Fatal("Execute() error = nil, want header validation")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), tt.want)
			}
			if strings.Contains(err.Error(), "yt auth login") {
				t.Fatalf("Execute() error = %q, parsed header after auth", err.Error())
			}
		})
	}
}

func TestParseRawHeadersProperties(t *testing.T) {
	property := func(value string) bool {
		headers, err := parseRawHeaders([]string{"X-Agent-Trace: " + value})
		return err == nil && headers.Get("X-Agent-Trace") == strings.TrimSpace(value)
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseRawQueryProperties(t *testing.T) {
	property := func(value string) bool {
		query, err := parseRawQuery([]string{"fields=" + value})
		return err == nil && len(query["fields"]) == 1 && query["fields"][0] == value
	}
	if err := quick.Check(property, nil); err != nil {
		t.Fatal(err)
	}
}

func FuzzParseRawHeaders(f *testing.F) {
	for _, seed := range []string{
		"X-Agent-Trace: run-1",
		"Accept: application/xml",
		"Authorization: Bearer bad",
		"Content-Type: text/plain",
		"missing-colon",
		": empty-name",
		"X-Colon: value:with:colon",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		headers, err := parseRawHeaders([]string{raw})
		name, _, hasColon := strings.Cut(raw, ":")
		name = strings.TrimSpace(name)
		managed := strings.EqualFold(name, "Authorization") || strings.EqualFold(name, "Content-Type")
		if !hasColon || name == "" || managed || !validHTTPHeaderName(name) {
			if err == nil {
				t.Fatalf("parseRawHeaders(%q) succeeded for invalid or managed header %#v", raw, headers)
			}
			return
		}
		if err != nil {
			t.Fatalf("parseRawHeaders(%q) error = %v", raw, err)
		}
		if len(headers) != 1 {
			t.Fatalf("parseRawHeaders(%q) produced %d headers, want 1", raw, len(headers))
		}
	})
}

func TestRawRejectsMultipleBodySources(t *testing.T) {
	err := Execute(context.Background(), []string{
		"raw", "/api/issues", "--body", `{}`, "--body-stdin",
	}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Execute() error = nil, want body source validation")
	}
	if !strings.Contains(err.Error(), "body accepts only one text source") {
		t.Fatalf("Execute() error = %q, want body source validation", err.Error())
	}
}
