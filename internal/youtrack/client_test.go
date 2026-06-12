package youtrack

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestArticlesRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/articles" {
			t.Fatalf("path = %s, want /api/articles", r.URL.Path)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "content") || !strings.Contains(fields, "idReadable") {
			t.Fatalf("fields = %q, want article fields", fields)
		}
		_, _ = w.Write([]byte(`[{"id":"226-1","idReadable":"ABC-A-1","summary":"Guide","project":{"shortName":"ABC"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	articles, err := client.Articles(context.Background(), ArticleListOptions{Top: 42})
	if err != nil {
		t.Fatalf("Articles() error = %v", err)
	}
	if len(articles) != 1 || articles[0].IDReadable != "ABC-A-1" {
		t.Fatalf("Articles() = %#v", articles)
	}
}

func TestProjectArticlesRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/projects/ABC/articles" {
			t.Fatalf("path = %s, want project articles path", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"226-1","idReadable":"ABC-A-1","summary":"Guide"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	articles, err := client.Articles(context.Background(), ArticleListOptions{Project: "ABC"})
	if err != nil {
		t.Fatalf("Articles() error = %v", err)
	}
	if len(articles) != 1 || articles[0].Summary != "Guide" {
		t.Fatalf("Articles() = %#v", articles)
	}
}

func TestAgilesRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agiles" {
			t.Fatalf("path = %s, want /api/agiles", r.URL.Path)
		}
		if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "currentSprint") {
			t.Fatalf("fields = %q, want agile fields", fields)
		}
		_, _ = w.Write([]byte(`[{"id":"120-1","name":"Team Board","currentSprint":{"id":"121-1","name":"Sprint 1"}}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	agiles, err := client.Agiles(context.Background(), PageOptions{Top: 42})
	if err != nil {
		t.Fatalf("Agiles() error = %v", err)
	}
	if len(agiles) != 1 || agiles[0].CurrentSprint.Name != "Sprint 1" {
		t.Fatalf("Agiles() = %#v", agiles)
	}
}

func TestSprintsRequestShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agiles/120-1/sprints" {
			t.Fatalf("path = %s, want sprints path", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"121-1","name":"Sprint 1"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	sprints, err := client.Sprints(context.Background(), SprintListOptions{AgileID: "120-1"})
	if err != nil {
		t.Fatalf("Sprints() error = %v", err)
	}
	if len(sprints) != 1 || sprints[0].Name != "Sprint 1" {
		t.Fatalf("Sprints() = %#v", sprints)
	}
}

func TestHelpdeskProjectsFilterProjectType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/projects" {
			t.Fatalf("path = %s, want /api/admin/projects", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[
			{"id":"0-1","shortName":"ABC","name":"Alpha","projectType":{"name":"standard"}},
			{"id":"0-2","shortName":"SUP","name":"Support","projectType":{"name":"helpdesk"}}
		]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	projects, err := client.HelpdeskProjects(context.Background(), PageOptions{})
	if err != nil {
		t.Fatalf("HelpdeskProjects() error = %v", err)
	}
	if len(projects) != 1 || projects[0].ShortName != "SUP" {
		t.Fatalf("HelpdeskProjects() = %#v", projects)
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

func TestAttachmentContentUsesThumbnailURL(t *testing.T) {
	const image = "png bytes"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtrack/api/files/preview.png" {
			t.Fatalf("path = %s, want thumbnail path", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		if got := r.Header.Get("Accept"); got != "*/*" {
			t.Fatalf("accept = %q, want bytes accept header", got)
		}
		_, _ = w.Write([]byte(image))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/youtrack", "perm:test", server.Client())
	data, err := client.AttachmentContent(context.Background(), AttachmentContentRequest{
		Attachment: Attachment{ThumbnailURL: "/youtrack/api/files/preview.png"},
		MaxBytes:   1024,
	})
	if err != nil {
		t.Fatalf("AttachmentContent() error = %v", err)
	}
	if string(data) != image {
		t.Fatalf("AttachmentContent() = %q, want image bytes", string(data))
	}
}

func TestFileContentDownloadsSameOriginBytes(t *testing.T) {
	const image = "image bytes"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtrack/api/files/image.png" {
			t.Fatalf("path = %s, want file path", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte(image))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/youtrack", "perm:test", server.Client())
	data, err := client.FileContent(context.Background(), FileContentRequest{
		URL:      "/youtrack/api/files/image.png",
		MaxBytes: 1024,
	})
	if err != nil {
		t.Fatalf("FileContent() error = %v", err)
	}
	if string(data) != image {
		t.Fatalf("FileContent() = %q, want image bytes", string(data))
	}
}

func TestAttachmentContentResolvesBaseRelativeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtrack/api/files/preview.png" {
			t.Fatalf("path = %s, want base-relative thumbnail path", r.URL.Path)
		}
		_, _ = w.Write([]byte("preview"))
	}))
	defer server.Close()

	client := NewClient(server.URL+"/youtrack", "perm:test", server.Client())
	_, err := client.AttachmentContent(context.Background(), AttachmentContentRequest{
		Attachment: Attachment{ThumbnailURL: "api/files/preview.png"},
		MaxBytes:   1024,
	})
	if err != nil {
		t.Fatalf("AttachmentContent() error = %v", err)
	}
}

func TestAttachmentContentRejectsCrossOriginURL(t *testing.T) {
	client := NewClient("https://youtrack.example.com", "perm:test", nil)
	_, err := client.AttachmentContent(context.Background(), AttachmentContentRequest{
		Attachment: Attachment{URL: "https://evil.example.com/file.png"},
	})
	if err == nil {
		t.Fatal("AttachmentContent() error = nil, want cross-origin rejection")
	}
	if !strings.Contains(err.Error(), "base origin") {
		t.Fatalf("AttachmentContent() error = %q, want base origin", err.Error())
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

func TestRawRequestContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/custom" {
			t.Fatalf("path = %s, want raw path", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "text/plain" {
			t.Fatalf("Content-Type = %q, want text/plain", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(body) != "plain text" {
			t.Fatalf("body = %q, want plain text", string(body))
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	data, err := client.Raw(context.Background(), RawRequest{
		Method:      http.MethodPatch,
		Path:        "/api/custom",
		ContentType: "text/plain",
		Body:        strings.NewReader("plain text"),
	})
	if err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("Raw() = %s, want raw response", data)
	}
}

func TestRawRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer perm:test" {
			t.Fatalf("Authorization = %q, want configured bearer token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/xml" {
			t.Fatalf("Accept = %q, want application/xml", got)
		}
		if got := r.Header.Get("X-Youtrack-Trace"); got != "agent-run-1" {
			t.Fatalf("X-YouTrack-Trace = %q, want agent-run-1", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	data, err := client.Raw(context.Background(), RawRequest{
		Method: http.MethodGet,
		Path:   "/api/custom",
		Headers: http.Header{
			"Accept":           {"application/xml"},
			"X-YouTrack-Trace": {"agent-run-1"},
		},
	})
	if err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("Raw() = %s, want raw response", data)
	}
}

func TestAddRawHeaderOwnsRawHeaderPolicy(t *testing.T) {
	headers := http.Header{}
	if err := AddRawHeader(headers, "x-youtrack-trace", " run-1 "); err != nil {
		t.Fatalf("AddRawHeader() error = %v", err)
	}
	if got := headers.Get("X-Youtrack-Trace"); got != "run-1" {
		t.Fatalf("header = %q, want trimmed value", got)
	}
	if err := AddRawHeader(headers, "Authorization", "Bearer bad"); err == nil {
		t.Fatal("AddRawHeader() error = nil, want managed authorization rejection")
	}
	if err := AddRawHeader(headers, "Content-Type", "text/plain"); err == nil {
		t.Fatal("AddRawHeader() error = nil, want managed content type rejection")
	}
	if err := AddRawHeader(headers, "Bad Header", "value"); err == nil {
		t.Fatal("AddRawHeader() error = nil, want header name validation")
	}
}

func TestRawRequestMergesQueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query()["existing"]; len(got) != 1 || got[0] != "true" {
			t.Fatalf("existing query = %#v, want true", got)
		}
		if got := r.URL.Query()["fields"]; len(got) != 1 || got[0] != "id,idReadable" {
			t.Fatalf("fields query = %#v, want id,idReadable", got)
		}
		if got := r.URL.Query()["tag"]; len(got) != 2 || got[0] != "agent" || got[1] != "urgent" {
			t.Fatalf("tag query = %#v, want repeated raw query values", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "perm:test", server.Client())
	data, err := client.Raw(context.Background(), RawRequest{
		Method: http.MethodGet,
		Path:   "/api/issues?existing=true",
		Query: url.Values{
			"fields": {"id,idReadable"},
			"tag":    {"agent", "urgent"},
		},
	})
	if err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("Raw() = %s, want raw response", data)
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
	if _, err := client.Raw(context.Background(), RawRequest{Path: "/api/issues"}); err == nil {
		t.Fatal("Raw() error = nil, want method validation")
	}
	if _, err := client.Raw(context.Background(), RawRequest{Method: http.MethodGet}); err == nil {
		t.Fatal("Raw() error = nil, want path validation")
	}
	if _, err := client.Raw(context.Background(), RawRequest{
		Method:  http.MethodGet,
		Path:    "/api/issues",
		Headers: http.Header{"Authorization": {"Bearer bad"}},
	}); err == nil {
		t.Fatal("Raw() error = nil, want managed authorization header validation")
	}
	if _, err := client.Raw(context.Background(), RawRequest{
		Method:  http.MethodGet,
		Path:    "/api/issues",
		Headers: http.Header{"Content-Type": {"text/plain"}},
	}); err == nil {
		t.Fatal("Raw() error = nil, want managed content type header validation")
	}
	if _, err := client.Raw(context.Background(), RawRequest{
		Method:  http.MethodGet,
		Path:    "/api/issues",
		Headers: http.Header{"Bad Header": {"value"}},
	}); err == nil {
		t.Fatal("Raw() error = nil, want header name validation")
	}
	if _, err := client.Raw(context.Background(), RawRequest{
		Method: http.MethodGet,
		Path:   "/api/issues",
		Query:  url.Values{"": {"value"}},
	}); err == nil {
		t.Fatal("Raw() error = nil, want query name validation")
	}
	if _, err := client.Projects(context.Background(), PageOptions{Top: -1}); err == nil {
		t.Fatal("Projects() error = nil, want top validation")
	}
	if _, err := client.Issues(context.Background(), IssueListOptions{Skip: -1}); err == nil {
		t.Fatal("Issues() error = nil, want skip validation")
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
