package youtrack

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

func TestWorkItemsRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/timeTracking/workItems" {
			t.Fatalf("path = %s, want work items path", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "10" {
			t.Fatalf("$top = %q", got)
		}
		_, _ = w.Write([]byte(`[{"id":"115-1","duration":{"minutes":30,"presentation":"30m"},"text":"review","author":{"login":"jane"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	items, err := client.WorkItems(context.Background(), WorkItemListOptions{IssueID: "ABC-1", Top: 10})
	if err != nil {
		t.Fatalf("WorkItems() error = %v", err)
	}
	if len(items) != 1 || items[0].Duration.Minutes != 30 {
		t.Fatalf("WorkItems() = %#v", items)
	}
}

func TestAddWorkItemRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/timeTracking/workItems" {
			t.Fatalf("path = %s, want work items path", r.URL.Path)
		}
		if got := r.URL.Query().Get("muteUpdateNotifications"); got != "true" {
			t.Fatalf("muteUpdateNotifications = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		duration := payload["duration"].(map[string]any)
		if duration["minutes"] != float64(45) {
			t.Fatalf("duration.minutes = %v", duration["minutes"])
		}
		if payload["text"] != "implementation" {
			t.Fatalf("text = %v", payload["text"])
		}
		workType := payload["type"].(map[string]any)
		if workType["id"] != "65-1" {
			t.Fatalf("type.id = %v", workType["id"])
		}
		_, _ = w.Write([]byte(`{"id":"115-1","duration":{"minutes":45,"presentation":"45m"},"text":"implementation","type":{"id":"65-1","name":"Development"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	item, err := client.AddWorkItem(context.Background(), AddWorkItemRequest{
		IssueID: "ABC-1",
		Minutes: 45,
		Text:    "implementation",
		TypeID:  "65-1",
		Mute:    true,
	})
	if err != nil {
		t.Fatalf("AddWorkItem() error = %v", err)
	}
	if item.Duration.Minutes != 45 {
		t.Fatalf("AddWorkItem() = %#v", item)
	}
}

func TestAttachmentsRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/attachments" {
			t.Fatalf("path = %s, want attachments path", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "10" {
			t.Fatalf("$top = %q", got)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "thumbnailURL") {
			t.Fatalf("fields = %q, want attachment fields", fields)
		}
		_, _ = w.Write([]byte(`[{"id":"134-1","name":"screenshot.png","size":123,"mimeType":"image/png","author":{"login":"jane"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	attachments, err := client.Attachments(context.Background(), AttachmentListOptions{IssueID: "ABC-1", Top: 10})
	if err != nil {
		t.Fatalf("Attachments() error = %v", err)
	}
	if len(attachments) != 1 || attachments[0].Name != "screenshot.png" {
		t.Fatalf("Attachments() = %#v", attachments)
	}
}

func TestUploadAttachmentsRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/attachments" {
			t.Fatalf("path = %s, want attachments path", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart form", r.Header.Get("Content-Type"))
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}
		parts := readMultipartParts(t, reader)
		if len(parts) != 2 {
			t.Fatalf("parts = %#v, want 2", parts)
		}
		if parts[0].field != "upload" || parts[0].filename != "one.txt" || parts[0].body != "one" {
			t.Fatalf("first part = %#v", parts[0])
		}
		if parts[1].field != "upload" || parts[1].filename != "two.txt" || parts[1].body != "two" {
			t.Fatalf("second part = %#v", parts[1])
		}
		_, _ = w.Write([]byte(`[{"id":"134-1","name":"one.txt"},{"id":"134-2","name":"two.txt"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	attachments, err := client.UploadAttachments(context.Background(), UploadAttachmentsRequest{
		IssueID: "ABC-1",
		Files: []AttachmentFile{
			{Name: "one.txt", Content: strings.NewReader("one")},
			{Name: "two.txt", Content: strings.NewReader("two")},
		},
	})
	if err != nil {
		t.Fatalf("UploadAttachments() error = %v", err)
	}
	if len(attachments) != 2 || attachments[1].Name != "two.txt" {
		t.Fatalf("UploadAttachments() = %#v", attachments)
	}
}

func TestActivitiesRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/activities" {
			t.Fatalf("path = %s, want activities path", r.URL.Path)
		}
		if got := r.URL.Query().Get("categories"); got != "CustomFieldCategory,CommentsCategory" {
			t.Fatalf("categories = %q", got)
		}
		if got := r.URL.Query().Get("reverse"); got != "true" {
			t.Fatalf("reverse = %q", got)
		}
		if got := r.URL.Query().Get("author"); got != "me" {
			t.Fatalf("author = %q", got)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "added") || !strings.Contains(fields, "removed") {
			t.Fatalf("fields = %q, want activity fields", fields)
		}
		_, _ = w.Write([]byte(`[{"id":"a-1","$type":"CustomFieldActivityItem","timestamp":1648110830229,"author":{"login":"jane"},"field":{"name":"State"},"added":[{"name":"Fixed"}],"removed":[{"name":"Open"}]}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	activities, err := client.Activities(context.Background(), ActivityListOptions{
		IssueID:    "ABC-1",
		Categories: []string{"CustomFieldCategory", "CommentsCategory"},
		Reverse:    true,
		Author:     "me",
	})
	if err != nil {
		t.Fatalf("Activities() error = %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("Activities() = %#v, want one activity", activities)
	}
	if summary := activities[0].Summary(); summary != "State +Fixed -Open" {
		t.Fatalf("Activity Summary() = %q", summary)
	}
}

func TestIssueLinksRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/ABC-1/links" {
			t.Fatalf("path = %s, want issue links path", r.URL.Path)
		}
		if got := r.URL.Query().Get("$top"); got != "10" {
			t.Fatalf("$top = %q, want 10", got)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "linkType") || !strings.Contains(fields, "trimmedIssues") {
			t.Fatalf("fields = %q, want issue link fields", fields)
		}
		_, _ = w.Write([]byte(`[{"id":"80-1s","direction":"OUTWARD","linkType":{"name":"Depend","sourceToTarget":"is required for","targetToSource":"depends on","directed":true},"issues":[{"id":"2-44","idReadable":"ABC-2","summary":"Linked issue"}]}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	links, err := client.IssueLinks(context.Background(), IssueLinkListOptions{
		IssueID: "ABC-1",
		Top:     10,
	})
	if err != nil {
		t.Fatalf("IssueLinks() error = %v", err)
	}
	if len(links) != 1 || links[0].Issues[0].IDReadable != "ABC-2" {
		t.Fatalf("IssueLinks() = %#v", links)
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
	if _, err := client.WorkItems(context.Background(), WorkItemListOptions{}); err == nil {
		t.Fatal("WorkItems() error = nil, want issue id validation")
	}
	if _, err := client.AddWorkItem(context.Background(), AddWorkItemRequest{IssueID: "ABC-1"}); err == nil {
		t.Fatal("AddWorkItem() error = nil, want minutes validation")
	}
	if _, err := client.Attachments(context.Background(), AttachmentListOptions{}); err == nil {
		t.Fatal("Attachments() error = nil, want issue id validation")
	}
	if _, err := client.UploadAttachments(context.Background(), UploadAttachmentsRequest{IssueID: "ABC-1"}); err == nil {
		t.Fatal("UploadAttachments() error = nil, want file validation")
	}
	if _, err := client.Activities(context.Background(), ActivityListOptions{}); err == nil {
		t.Fatal("Activities() error = nil, want issue id validation")
	}
	if _, err := client.Activities(context.Background(), ActivityListOptions{IssueID: "ABC-1"}); err == nil {
		t.Fatal("Activities() error = nil, want category validation")
	}
	if _, err := client.IssueLinks(context.Background(), IssueLinkListOptions{}); err == nil {
		t.Fatal("IssueLinks() error = nil, want issue id validation")
	}
}

type multipartPart struct {
	field    string
	filename string
	body     string
}

func readMultipartParts(t *testing.T, reader *multipart.Reader) []multipartPart {
	t.Helper()

	var parts []multipartPart
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("next multipart part: %v", err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("read multipart part: %v", err)
		}
		parts = append(parts, multipartPart{
			field:    part.FormName(),
			filename: part.FileName(),
			body:     string(body),
		})
	}
	return parts
}
