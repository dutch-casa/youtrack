package youtrack

import (
	"context"
	"encoding/json"
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
	if _, err := client.AddComment(context.Background(), "ABC-1", ""); err == nil {
		t.Fatal("AddComment() error = nil, want text validation")
	}
}
