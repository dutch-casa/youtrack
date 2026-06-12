package youtrack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIssuesRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			t.Fatalf("path = %s, want /api/issues", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("Authorization = %q", got)
		}
		if got := r.URL.Query().Get("query"); got != "project: ABC" {
			t.Fatalf("query = %q", got)
		}
		if got := r.URL.Query().Get("$top"); got != "10" {
			t.Fatalf("$top = %q", got)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "idReadable") {
			t.Fatalf("fields = %q, want issue fields", fields)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"1","idReadable":"ABC-1","summary":"Test","project":{"shortName":"ABC"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	issues, err := client.Issues(context.Background(), IssueListOptions{Query: "project: ABC", Top: 10})
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}
	if len(issues) != 1 || issues[0].IDReadable != "ABC-1" {
		t.Fatalf("Issues() = %#v", issues)
	}
}

func TestCreateIssueRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		project := payload["project"].(map[string]any)
		if project["shortName"] != "ABC" {
			t.Fatalf("project shortName = %v", project["shortName"])
		}
		if payload["summary"] != "New issue" {
			t.Fatalf("summary = %v", payload["summary"])
		}
		_, _ = w.Write([]byte(`{"id":"1","idReadable":"ABC-1","summary":"New issue","project":{"shortName":"ABC"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	issue, err := client.CreateIssue(context.Background(), CreateIssueRequest{ProjectShortName: "ABC", Summary: "New issue"})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}
	if issue.IDReadable != "ABC-1" {
		t.Fatalf("CreateIssue() = %#v", issue)
	}
}

func TestUpdateIssueRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1" {
			t.Fatalf("path = %s, want /api/issues/ABC-1", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["summary"] != "Updated" {
			t.Fatalf("summary = %v", payload["summary"])
		}
		if _, ok := payload["description"]; ok {
			t.Fatalf("description was sent unexpectedly: %#v", payload)
		}
		_, _ = w.Write([]byte(`{"id":"1","idReadable":"ABC-1","summary":"Updated","project":{"shortName":"ABC"}}`))
	}))
	defer server.Close()

	summary := "Updated"
	client := NewClient(server.URL, "perm:test", server.Client())
	issue, err := client.UpdateIssue(context.Background(), UpdateIssueRequest{ID: "ABC-1", Summary: &summary})
	if err != nil {
		t.Fatalf("UpdateIssue() error = %v", err)
	}
	if issue.Summary != "Updated" {
		t.Fatalf("UpdateIssue() = %#v", issue)
	}
}

func TestProjectsRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/projects" {
			t.Fatalf("path = %s, want /api/admin/projects", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "42" {
			t.Fatalf("$top = %q", got)
		}
		_, _ = w.Write([]byte(`[{"id":"0-1","shortName":"ABC","name":"Alpha"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	projects, err := client.Projects(context.Background(), PageOptions{Top: 42})
	if err != nil {
		t.Fatalf("Projects() error = %v", err)
	}
	if len(projects) != 1 || projects[0].ShortName != "ABC" {
		t.Fatalf("Projects() = %#v", projects)
	}
}

func TestUsersRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Fatalf("path = %s, want /api/users", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"1-1","login":"jane","name":"Jane"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	users, err := client.Users(context.Background(), PageOptions{})
	if err != nil {
		t.Fatalf("Users() error = %v", err)
	}
	if len(users) != 1 || users[0].Login != "jane" {
		t.Fatalf("Users() = %#v", users)
	}
}

func TestApplyCommandRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/commands" {
			t.Fatalf("path = %s, want /api/commands", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload["query"] != "State Fixed" {
			t.Fatalf("query = %v", payload["query"])
		}
		issues := payload["issues"].([]any)
		issue := issues[0].(map[string]any)
		if issue["idReadable"] != "ABC-1" {
			t.Fatalf("issue idReadable = %v", issue["idReadable"])
		}
		if payload["comment"] != "done" {
			t.Fatalf("comment = %v", payload["comment"])
		}
		_, _ = w.Write([]byte(`{"query":"State Fixed","issues":[{"id":"1","idReadable":"ABC-1","summary":"Done"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	result, err := client.ApplyCommand(context.Background(), ApplyCommandRequest{
		IssueID: "ABC-1",
		Query:   "State Fixed",
		Comment: "done",
		Silent:  true,
	})
	if err != nil {
		t.Fatalf("ApplyCommand() error = %v", err)
	}
	if result.Query != "State Fixed" || len(result.Issues) != 1 {
		t.Fatalf("ApplyCommand() = %#v", result)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error_description":"bad token"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad", server.Client())
	_, err := client.CurrentUser(context.Background())
	if err == nil {
		t.Fatal("CurrentUser() error = nil, want APIError")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("CurrentUser() error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestClientValidatesInputs(t *testing.T) {
	client := NewClient("https://example.youtrack.cloud", "perm:test", nil)

	if _, err := client.Issue(context.Background(), ""); err == nil {
		t.Fatal("Issue() error = nil, want issue id validation")
	}
	if _, err := client.CreateIssue(context.Background(), CreateIssueRequest{}); err == nil {
		t.Fatal("CreateIssue() error = nil, want request validation")
	}
	if _, err := client.UpdateIssue(context.Background(), UpdateIssueRequest{ID: "ABC-1"}); err == nil {
		t.Fatal("UpdateIssue() error = nil, want field validation")
	}
	if _, err := client.AddComment(context.Background(), "ABC-1", ""); err == nil {
		t.Fatal("AddComment() error = nil, want text validation")
	}
	if _, err := client.ApplyCommand(context.Background(), ApplyCommandRequest{IssueID: "ABC-1"}); err == nil {
		t.Fatal("ApplyCommand() error = nil, want query validation")
	}
}
